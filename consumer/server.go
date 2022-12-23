package consumer

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/senpan/xserver/bootstrap"
	"github.com/senpan/xserver/consumer/core"
	"github.com/senpan/xserver/consumer/internal/ctrl"
	"github.com/senpan/xserver/logger"
)

type MQHandler func(topic string, data []byte, other ...string) error

type MQConsumerServer struct {
	funcSetter *bootstrap.FuncSetter
	handlers   map[string]core.MQHandler
	ctl        *ctrl.Ctrl
	path       string
	exit       chan struct{}
}

// NewMQConsumerServer creates a new MQ consumer server
// @param path string 配置文件路径
func NewMQConsumerServer(path string) *MQConsumerServer {
	srv := new(MQConsumerServer)
	srv.funcSetter = bootstrap.NewFuncSetter()
	srv.path = path
	srv.exit = make(chan struct{})
	return srv
}
func (s *MQConsumerServer) Serve() (err error) {
	if err = s.funcSetter.RunStartFunc(); err != nil {
		return nil
	}

	go s.regSignal()
	go s.doJob()

	<-s.exit
	logger.GetLogger().Warn("xserver.MQConsumerServer", "Stop Complete.")
	return nil
}

func (s *MQConsumerServer) waitShutdown() {
	logger.GetLogger().Warn("xserver.MQConsumerServer.Stop", "Process Stop...")

	s.ctl.Close()
	s.funcSetter.RunStopFunc()
	logger.GetLogger().Warn("xserver.MQConsumerServer.Stop", "Process Stop Complete")
	time.Sleep(1 * time.Second)
	s.exit <- struct{}{}
}

func (s *MQConsumerServer) AddHandler(name string, fn core.MQHandler) {
	s.handlers[name] = fn
}

func (s *MQConsumerServer) regSignal() {
	sg := make(chan os.Signal, 2)
	signal.Notify(sg, os.Interrupt, syscall.SIGTERM)
	<-sg
	s.waitShutdown()
}

func (s *MQConsumerServer) doJob() {
	c := ctrl.NewCtrl(s.path, s.handlers)
	s.ctl = c
}

func (s *MQConsumerServer) AddStartFunc(fns ...bootstrap.ServerStartFunc) {
	s.funcSetter.AddStartFunc(fns...)
}

func (s *MQConsumerServer) AddStopFunc(fns ...bootstrap.ServerStopFunc) {
	s.funcSetter.AddStopFunc(fns...)
}
