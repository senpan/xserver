package consumers

import (
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/spf13/cast"

	logger "github.com/senpan/xlogger"

	"github.com/senpan/xserver/consumer/core"
)

var produce sarama.SyncProducer
var once = new(sync.Once)

type KafkaConsumer struct {
	exit         chan struct{}
	callback     core.MQHandler
	successCount int64
	errorCount   int64
	wg           *sync.WaitGroup
}

type KafkaConfig struct {
	Host          []string `json:"host"`
	FailHost      []string `json:"failHost"`
	Topic         string   `json:"topic"`
	SASL          *SASL    `json:"sasl"`
	FailTopic     string   `json:"failTopic"`
	FailCount     int      `json:"failCount"`
	ConsumerGroup string   `json:"consumerGroup"`
	ConsumerCount int      `json:"consumerCount"`
	Handler       string   `json:"handler"`
}

type SASL struct {
	Enabled  bool   `json:"enabled"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func NewKafkaConsumer(configs []*KafkaConfig, handlers map[string]core.MQHandler) (consumer *KafkaConsumer, err error) {
	tag := "xserver.consumer.kafka"
	if len(configs) == 0 {
		err = logger.NewError("kafka config not found")
		logger.E(tag, "kafka config not found")
		return
	}
	consumer = new(KafkaConsumer)

	consumer.exit = make(chan struct{})
	consumer.wg = new(sync.WaitGroup)
	go consumer.count()
	for _, config := range configs {
		if h, ok := handlers[config.Handler]; !ok {
			logger.E(tag, "topic:%s, not found mq handler", config.Topic)
			continue
		} else {
			consumer.callback = h
		}
		for i := 0; i < config.ConsumerCount; i++ {
			consumer.wg.Add(1)
			go consumer.consume(config)
		}
	}
	return
}

func (k *KafkaConsumer) Close() {
	tag := "xserver.consumer.kafka"
	logger.W(tag, "kafka,close consumers begin")
	close(k.exit)
	done := make(chan struct{})
	go func() {
		k.wg.Wait()
		done <- struct{}{}
	}()
	select {
	case <-done:
		logger.W(tag, "kafka,close consumers wait for done")
	case <-time.After(time.Second * 2):
		logger.E(tag, "kafka,close consumers wait timeout")
	}
}

func (k *KafkaConsumer) count() {
	tag := "xserver.consumer.stat"
	t := time.NewTicker(time.Second * 60)
	for {
		select {
		case <-t.C:
			succ := atomic.SwapInt64(&k.successCount, 0)
			fail := atomic.SwapInt64(&k.errorCount, 0)
			logger.I(tag, "[Kafka Stat] success:%d,fail:%d", succ, fail)
		}
	}
}

func (k *KafkaConsumer) initSarama(kfg *KafkaConfig) (cs *cluster.Consumer) {
	tag := "xserver.consumer.kafka"
	cfg := cluster.NewConfig()
	cfg.Config.ClientID = "xconsumer.kafkaWorker"
	if kfg.SASL.Enabled {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.User = kfg.SASL.User
		cfg.Net.SASL.Password = kfg.SASL.Password
	}

	cfg.Config.Consumer.MaxWaitTime = 500 * time.Millisecond
	cfg.Config.Consumer.MaxProcessingTime = 300 * time.Millisecond
	cfg.Config.Consumer.Offsets.AutoCommit.Enable = true
	cfg.Config.Consumer.Offsets.AutoCommit.Interval = 350 * time.Millisecond
	cfg.Config.Consumer.Offsets.Initial = sarama.OffsetNewest
	cfg.Config.Consumer.Offsets.Retention = time.Hour * 24 * 15
	cfg.Config.Consumer.Return.Errors = true
	cfg.Group.Return.Notifications = true
	cfg.Version = sarama.V0_11_0_2
	cs, err := cluster.NewConsumer(kfg.Host, kfg.ConsumerGroup, strings.Split(kfg.Topic, ","), cfg)
	if err != nil {
		logger.F(tag, "newConsumer err:%v", err)
	}
	if kfg.FailTopic != "" {
		once.Do(func() {
			config := sarama.NewConfig()
			config.Producer.Partitioner = sarama.NewHashPartitioner
			config.Producer.RequiredAcks = sarama.WaitForAll
			config.Producer.Return.Successes = true
			if kfg.SASL.Enabled {
				config.Net.SASL.Enable = true
				config.Net.SASL.User = kfg.SASL.User
				config.Net.SASL.Password = kfg.SASL.Password
			}
			hosts := kfg.Host
			if len(kfg.FailHost) > 0 {
				hosts = kfg.FailHost
			}
			logger.D(tag, "NewSyncProducer hosts:%v", hosts)
			if produce, err = sarama.NewSyncProducer(hosts, config); err != nil {
				logger.F(tag, "NewSyncProducer err:%v", err)
			}
		})
	}
	return
}

func (k *KafkaConsumer) consume(config *KafkaConfig) {
	tag := "xserver.consumer.kafka"
	defer k.wg.Done()
	defer k.recovery()
	logger.I(tag, "Start consumer from broker %v", config.Host)
	// 初始化消费队列
	cs := k.initSarama(config)
	if cs != nil {
		defer cs.Close()
	}

	go func(c *cluster.Consumer) {
		for notification := range c.Notifications() {
			logger.D(tag, "ReBalance:%+v", notification)
		}
	}(cs)
	go func(c *cluster.Consumer) {
		for err := range c.Errors() {
			logger.E(tag, "consumer errors,err:%v", err)
		}
	}(cs)
	for {
		message := cs.Messages()
		select {
		case <-k.exit:
			logger.W(tag, "[Quit] accept quit signal")
			if err := cs.CommitOffsets(); err != nil {
				logger.E(tag, "[Quit] commit offset error:%v", err)
				time.Sleep(time.Second * 2)
			}
			return
		case event, ok := <-message:
			if !ok {
				continue
			}
			count := 0
			for {
				ret := k.callback(config.Topic, event.Value, string(event.Key))
				if ret == nil {
					atomic.AddInt64(&k.successCount, int64(1))
					logger.D(tag, "[Consumer] success,key:%s,offset:%d,partition:%d", string(event.Key), event.Offset, event.Partition)
					cs.MarkOffset(event, "")
					break
				} else {
					logger.E(tag, "[Consumer] failed:key:%s,val:%s,offset:%d,partition:%d", string(event.Key), string(event.Value), event.Offset, event.Partition)
					if count >= config.FailCount {
						if config.FailTopic != "" {
							for i := 0; i < 3; i++ {
								if err := k.sendByHashPartition(config.FailTopic, event.Value, event.Value); err != nil {
									logger.E(tag, "[Consumer] send to fail topic,topic:%s,val:%s,err:%v", config.FailTopic, string(event.Value), err)
									continue
								}
								break
							}
							logger.D(tag, "[Consumer] send to fail topic,key:%s,val:%s,offset:%d,partition:%d", string(event.Key), string(event.Value), event.Offset, event.Partition)
							cs.MarkOffset(event, "")
							break
						} else {
							logger.E(tag, "[Consumer] unset fail topic:key:%s,val:%s,offset:%d,partition:%d", string(event.Key), string(event.Value), event.Offset, event.Partition)
							break
						}
					}
					count++
				}
			}
		}
	}
}

func (k *KafkaConsumer) sendByHashPartition(topic string, data []byte, key []byte) (errr error) {
	tag := "xserver.consumer.kafka"
	msg := &sarama.ProducerMessage{Topic: topic, Key: sarama.ByteEncoder(key), Value: sarama.ByteEncoder(data)}
	if partition, offset, err := produce.SendMessage(msg); err != nil {
		errr = err
	} else {
		logger.D(tag, "[Fail Topic] sendByHashPartition,topic:%s,partition:%d,offset:%d,data:%s", topic, partition, offset, string(data))
	}
	return
}

func (k *KafkaConsumer) recovery() {
	tag := "xserver.consumer.kafka"
	if rec := recover(); rec != nil {
		if err, ok := rec.(error); ok {
			logger.E(tag, "[Recovery] Unhandled error: %v\n stack:%v", err.Error(), cast.ToString(debug.Stack()))
		} else {
			logger.E(tag, "[Recovery] Panic: %v\n stack:%v", rec, cast.ToString(debug.Stack()))
		}
	}
}
