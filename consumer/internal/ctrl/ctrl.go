package ctrl

import (
	"github.com/senpan/xserver/consumer/core"
	"github.com/senpan/xserver/consumer/internal/consumers"
)

type Ctrl struct {
	consumers *consumers.ConsumerManager
}

func NewCtrl(path string, handlers map[string]core.MQHandler) *Ctrl {
	c := new(Ctrl)
	c.consumers = consumers.NewConsumerManager(path, handlers)
	return c
}

func (c *Ctrl) Close() {
	c.consumers.Close()
}
