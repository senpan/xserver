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

	logger "github.com/senpan/xlogger"
	"github.com/senpan/xtools/confx"

	"github.com/senpan/xserver/bootstrap"
)

type Server struct {
	funcSetter *bootstrap.FuncSetter
	server     *http.Server
	opts       ServerOptions
	exit       chan os.Signal
}

func NewServer() *Server {
	opts := DefaultOptions()
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
	err = s.funcSetter.RunServerStartFunc()
	if err != nil {
		return err
	}

	if s.opts.Grace && runtime.GOOS != "windows" {
		graceStart(s.opts.Addr, s.server)
	} else {
		signal.Notify(s.exit, os.Interrupt, syscall.SIGTERM)
		go s.waitShutdown()
		logger.I(tag, "http server start. Please visit %s", formatSvcAddr(s.server.Addr))
		err = s.server.ListenAndServe()
	}

	s.funcSetter.RunServerStopFunc()

	return err
}

// Shutdown close http server
func (s *Server) waitShutdown() {
	tag := "xserver.GinServer.waitShutdown"
	<-s.exit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.I(tag, "shutdown http server ...")

	err := s.server.Shutdown(ctx)
	if err != nil {
		logger.E(tag, "shutdown http server error:%+v", err)
	}
	return
}

func (s *Server) GetServer() *http.Server {
	return s.server
}

func (s *Server) GetGinEngine() *gin.Engine {
	return s.server.Handler.(*gin.Engine)
}

func (s *Server) AddServerStartFunc(fns ...bootstrap.ServerStartFunc) {
	s.funcSetter.AddServerStartFunc(fns...)
}

func (s *Server) AddServerStopFunc(fns ...bootstrap.ServerStopFunc) {
	s.funcSetter.AddServerStopFunc(fns...)
}

func (s *Server) InitConfig() bootstrap.ServerStartFunc {
	return func() error {
		err := confx.ParseConfToStruct("Server", &s.opts)
		if err != nil {
			return err
		}
		s.server.Addr = s.opts.Addr
		s.server.ReadTimeout = s.opts.ReadTimeout
		s.server.WriteTimeout = s.opts.WriteTimeout
		s.server.IdleTimeout = s.opts.IdleTimeout
		if s.opts.Mode != "" {
			_ = os.Setenv(gin.EnvGinMode, s.opts.Mode)
			gin.SetMode(s.opts.Mode)
		}
		return nil
	}
}

func formatSvcAddr(addr string) string {
	if addr[0:1] == ":" {
		return "127.0.0.1" + addr
	}
	return addr
}
