package ginhttp

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/senpan/xserver/bootstrap"
	"github.com/senpan/xserver/logger"
)

type Server struct {
	funcSetter *bootstrap.FuncSetter
	server     *http.Server
	opts       Options
	exit       chan os.Signal
}

func NewServer(options ...OptionFunc) *Server {
	opts := DefaultOptions()

	for _, o := range options {
		o(&opts)
	}

	s := new(Server)
	s.opts = opts
	s.funcSetter = bootstrap.NewFuncSetter()
	handler := gin.New()
	sev := &http.Server{
		Addr:         opts.Addr,
		Handler:      handler,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		IdleTimeout:  opts.IdleTimeout,
	}
	s.exit = make(chan os.Signal, 2)
	s.server = sev
	return s
}

// Serve serve http request
func (s *Server) Serve() error {
	tag := "xserver.GinServer.Serve"
	var err error
	err = s.funcSetter.RunStartFunc()
	if err != nil {
		return err
	}

	if s.opts.Mode != "" {
		_ = os.Setenv(gin.EnvGinMode, s.opts.Mode)
		gin.SetMode(s.opts.Mode)
	}

	if s.opts.Grace && runtime.GOOS != "windows" {
		graceStart(s.opts.Addr, s.server)
	} else {
		signal.Notify(s.exit, os.Interrupt, syscall.SIGTERM)
		go s.waitShutdown()
		logger.GetLogger().Infof(tag, "http server start. Please visit %s\n", formatSvcAddr(s.server.Addr))
		err = s.server.ListenAndServe()
	}

	s.funcSetter.RunStopFunc()

	return err
}

// Shutdown close http server
func (s *Server) waitShutdown() {
	tag := "xserver.GinServer.waitShutdown"
	<-s.exit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.GetLogger().Infof(tag, "shutdown http server ...")

	err := s.server.Shutdown(ctx)
	if err != nil {
		logger.GetLogger().Errorf(tag, "shutdown http server error:%+v", err)
	}
}

func (s *Server) GetServer() *http.Server {
	return s.server
}

func (s *Server) GetGinEngine() *gin.Engine {
	return s.server.Handler.(*gin.Engine)
}

func (s *Server) AddStartFunc(fns ...bootstrap.ServerStartFunc) {
	s.funcSetter.AddStartFunc(fns...)
}

func (s *Server) AddStopFunc(fns ...bootstrap.ServerStopFunc) {
	s.funcSetter.AddStopFunc(fns...)
}

func formatSvcAddr(addr string) string {
	if addr[0:1] == ":" {
		return "127.0.0.1" + addr
	}
	return addr
}
