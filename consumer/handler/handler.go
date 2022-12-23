package handler

import (
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/senpan/xserver/logger"
)

type MQHandler struct {
	Name string // 名称
	Task TaskFunc
}

type TaskFunc func(topic string, data []byte, other ...string) error

func (m *MQHandler) Do(topic string, data []byte, other ...string) (err error) {
	if m.Task == nil {
		return errors.New("task's function not defined")
	}
	// catch panic
	defer func() {
		if r := recover(); r != nil {
			logger.GetLogger().Errorf("MQHandler.Recovery", "err: %v, stacks: %s", r, string(debug.Stack()))
			err = fmt.Errorf("handler:%s,err: %v", m.Name, r)
			return
		}
	}()
	err = m.Task(topic, data, other...)
	return
}
