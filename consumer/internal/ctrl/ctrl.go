package ctrl

import (
	"github.com/senpan/xserver/consumer/handler"
	"github.com/senpan/xserver/consumer/internal/consumers"
)

type Ctrl struct {
	consumers *consumers.ConsumerManager
}

// NewCtrl 控制器
func NewCtrl(path string, handlers map[string]handler.MQHandler) *Ctrl {
	c := new(Ctrl)
	c.consumers = consumers.NewConsumerManager(path, handlers)
	return c
}

func (c *Ctrl) Close() {
	c.consumers.Close()
}
