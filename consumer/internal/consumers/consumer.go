package consumers

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/senpan/xserver/consumer/core"
	"github.com/senpan/xserver/logger"
)

type ConsumerManager struct {
	cfg      *Configure
	kafka    *KafkaConsumer
	rocketmq *RocketMQConsumer
}

type MQ struct {
	Kafka    bool `json:"kafka"`
	Rocketmq bool `json:"rocketmq"`
}

type Configure struct {
	Enabled  *MQ               `json:"enabled"`
	Kafka    []*KafkaConfig    `json:"kafka"`
	Rocketmq []*RocketMQConfig `json:"rocketmq"`
}

func NewConsumerManager(path string, handlers map[string]core.MQHandler) *ConsumerManager {
	tag := "xserver.consumer.consumerManager"
	var err error
	manager := new(ConsumerManager)
	cfg := loadMQConfig(path)
	manager.cfg = cfg
	if cfg.Enabled.Kafka {
		manager.kafka, err = NewKafkaConsumer(cfg.Kafka, handlers)
		if err != nil {
			logger.GetLogger().Errorf(tag, "NewKafkaConsumer  init err:%v", tag, err)
		}
	}
	if cfg.Enabled.Rocketmq {
		manager.rocketmq, err = NewRocketMQConsumer(cfg.Rocketmq, handlers)
		if err != nil {
			logger.GetLogger().Errorf(tag, "NewRocketConsumer  init err:%v", err)
		}
	}
	return manager
}

func (c *ConsumerManager) Close() {
	if c.cfg.Enabled.Kafka {
		c.kafka.Close()
	}
	if c.cfg.Enabled.Rocketmq {
		c.rocketmq.Close()
	}
}

// Load Config.json
func loadMQConfig(path string) *Configure {
	tag := "xserver.consumer.consumerManager"
	if path == "" {
		curPwd, _ := os.Getwd()
		path = filepath.Join(curPwd, "conf/consumer/mq.json")
	} else {
		path = filepath.Join(path, "mq.json")
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		logger.GetLogger().Fatalf(tag, "loadMQConfig:read file err:%v", err)
	}
	config := new(Configure)
	err = json.Unmarshal(b, config)
	if err != nil {
		logger.GetLogger().Fatalf(tag, "loadMQConfig:json Unmarshal err:%v", err)
	}
	return config
}
