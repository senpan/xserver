package consumer

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	logger "github.com/senpan/xlogger"

	"github.com/senpan/xserver/bootstrap"
	"github.com/senpan/xserver/consumer/core"
	"github.com/senpan/xserver/consumer/internal/ctrl"
)

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
	if err = s.funcSetter.RunServerStartFunc(); err != nil {
		return nil
	}

	go s.regSignal()
	go s.doJob()

	<-s.exit
	logger.W("xserver.MQConsumerServer", "Stop Complete.")
	return nil
}

func (s *MQConsumerServer) waitShutdown() {
	logger.W("xserver.MQConsumerServer.Stop", "Process Stop...")
	s.ctl.Close()
	s.funcSetter.RunServerStopFunc()
	logger.W("xserver.MQConsumerServer.Stop", "Process Stop Complete")
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

func (s *MQConsumerServer) AddServerStartFunc(fns ...bootstrap.ServerStartFunc) {
	s.funcSetter.AddServerStartFunc(fns...)
}

func (s *MQConsumerServer) AddServerStopFunc(fns ...bootstrap.ServerStopFunc) {
	s.funcSetter.AddServerStopFunc(fns...)
}
