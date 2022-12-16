package consumers

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"github.com/spf13/cast"

	logger "github.com/senpan/xlogger"

	"github.com/senpan/xserver/consumer/core"
	"github.com/senpan/xserver/consumer/internal/log"
)

type RocketMQConsumer struct {
	exit     chan struct{}
	callback core.MQHandler
	wg       *sync.WaitGroup
}

const (
	PushMode = iota
	PullMode
)

type RocketMQConfig struct {
	ConsumerMode  int                       `json:"consumerMode"` //0:push 1:pull
	ConsumerGroup string                    `json:"consumerGroup"`
	ConsumerCount int                       `json:"consumerCount"`
	NameServer    []string                  `json:"nameServer"`
	Namespace     string                    `json:"namespace"`
	Credentials   *RocketMQCredentials      `json:"credentials"`
	Topic         string                    `json:"topic"`
	Tags          string                    `json:"tags"`
	Retry         int                       `json:"retry"`
	Offset        consumer.ConsumeFromWhere `json:"offset"`
	Handler       string                    `json:"handler"`
}

type RocketMQCredentials struct {
	AccessKey     string `json:"accessKey"`
	SecretKey     string `json:"secretKey"`
	SecurityToken string `json:"securityToken"`
}

func NewRocketMQConsumer(configs []*RocketMQConfig, handlers map[string]core.MQHandler) (consumer *RocketMQConsumer, err error) {
	tag := "xserver.consumer.rocketmq"
	if len(configs) == 0 {
		err = logger.NewError("rocketmq config not found")
		return
	}
	consumer = new(RocketMQConsumer)
	consumer.exit = make(chan struct{})
	consumer.wg = new(sync.WaitGroup)

	rlog.SetLogger(&log.RocketMQLog{})

	for _, config := range configs {
		if h, ok := handlers[config.Handler]; !ok {
			logger.E(tag, "topic:%s, not found mq handler", config.Topic)
			continue
		} else {
			consumer.callback = h
		}
		go consumer.run(config)
	}
	return
}

func (r *RocketMQConsumer) Close() {
	tag := "xserver.consumer.rocketmq"
	logger.W(tag, "starting close consumers")
	close(r.exit)
	r.wg.Wait()
	logger.W(tag, "close consumers done")
}

func (r *RocketMQConsumer) pushConsumer(kfg *RocketMQConfig) {
	tag := "xserver.consumer.rocketmq"

	defer r.wg.Done()

	c, err := rocketmq.NewPushConsumer(
		consumer.WithInstance(r.getInstanceName(kfg)),
		consumer.WithGroupName(kfg.ConsumerGroup),
		consumer.WithNameServer(kfg.NameServer),
		consumer.WithRetry(kfg.Retry),
		consumer.WithConsumeFromWhere(kfg.Offset),
		consumer.WithCredentials(primitive.Credentials{
			AccessKey: kfg.Credentials.AccessKey,
			SecretKey: kfg.Credentials.SecretKey,
		}),
		consumer.WithNamespace(kfg.Namespace),
	)

	if err != nil {
		logger.E(tag, "[Push Consumer] new push consumers error:%v", err)
		return
	}

	ms := new(consumer.MessageSelector)
	ms.Type = consumer.TAG
	ms.Expression = kfg.Tags

	err = c.Subscribe(kfg.Topic, *ms, func(ctx context.Context,
		messages ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, me := range messages {
			for {
				ret := r.callback(me.Topic, me.Body, me.GetTags())
				if ret == nil {
					logger.D(tag, "[Push Consumer] success,topic:%s,tags:%s,handler:%s", me.Topic, me.GetTags(), kfg.Handler)
					break
				} else {
					logger.E(tag, "[Push Consumer] failed,topic:%s,tags:%s,val:%s", me.Topic, me.GetTags(), string(me.Body))
					// TODO:retry and send fail topic
					break
				}
			}
		}
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		logger.E(tag, "[Push Consumer] subscribe error:%v", err)
		return
	}
	err = c.Start()
	if err != nil {
		logger.E(tag, "[Push Consumer] start error:%v", err)
		return
	}
	for {
		select {
		case <-r.exit:
			logger.W(tag, "[Push Consumer] accept quit signal")
			return
		}
	}
}

func (r *RocketMQConsumer) pullConsumer(kfg *RocketMQConfig) {
	// TODO:support pull consumers
}

func (r *RocketMQConsumer) run(config *RocketMQConfig) {
	defer r.recovery()

	for i := 0; i < config.ConsumerCount; i++ {
		r.wg.Add(1)
		go r.consume(config, i)
	}
}

func (r *RocketMQConsumer) consume(config *RocketMQConfig, count int) {
	if config.ConsumerMode == PushMode {
		go r.pushConsumer(config)
	}
	if config.ConsumerMode == PullMode {
		go r.pullConsumer(config)
	}
}

// 自定义实例名称,防止出现Client ID 冲突问题
func (r *RocketMQConsumer) getInstanceName(kfg *RocketMQConfig) string {
	name := kfg.ConsumerGroup + "@"
	if kfg.Namespace != "" {
		name += kfg.Namespace
	} else {
		name += kfg.Topic
	}
	return name
}

func (r *RocketMQConsumer) recovery() {
	tag := "xserver.consumer.rocketmq"
	if rec := recover(); rec != nil {
		if err, ok := rec.(error); ok {
			logger.E(tag, "[Recovery] Unhandled error: %+v\n stack:%s", err.Error(), cast.ToString(debug.Stack()))
		} else {
			logger.E(tag, "[Recovery] Panic: %+v\n stack:%s", rec, cast.ToString(debug.Stack()))
		}
	}
}
