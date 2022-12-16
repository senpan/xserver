package clijob

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	logger "github.com/senpan/xlogger"
	"github.com/senpan/xtools/confx"

	"github.com/senpan/xserver/bootstrap"
)

type JobServer struct {
	funcSetter *bootstrap.FuncSetter
	opts       Options
	jobs       map[string]JobHandler
	cmdParser  CmdParser
	exit       chan struct{} // 退出
}

func init() {
	confx.InitConfig()
}

func NewJobServer(options ...OptionFunc) *JobServer {
	opts := DefaultOptions()

	for _, o := range options {
		o(&opts)
	}

	srv := &JobServer{
		funcSetter: bootstrap.NewFuncSetter(),
		opts:       opts,
		jobs:       make(map[string]JobHandler),
		exit:       make(chan struct{}),
		cmdParser:  opts.cmdParser,
	}

	return srv
}

func (js *JobServer) Serve() (err error) {

	if err = js.funcSetter.RunServerStartFunc(); err != nil {
		return nil
	}

	go js.regSignal()
	go js.doJob()

	<-js.exit
	return nil
}

func (js *JobServer) waitShutdown() {
	logger.I("xserver.JobServer.Stop", "Stop...")

	js.funcSetter.RunServerStopFunc()

	time.Sleep(1 * time.Second)

	js.exit <- struct{}{}
}

func (js *JobServer) regSignal() {
	sg := make(chan os.Signal, 2)
	signal.Notify(sg, os.Interrupt, syscall.SIGTERM)
	<-sg

	js.waitShutdown()
}

// AddHandler  添加执行脚本
func (js *JobServer) AddHandler(name string, fn TaskFunc) {
	js.jobs[name] = JobHandler{
		Name: name,
		Task: fn,
	}
}

func (js *JobServer) doJob() {
	tag := "xserver.JobServer.doJob"

	defer js.waitShutdown()
	defer processMark(tag, "JobServer Process")()

	jobSelected, err := js.cmdParser.JobArgParse(js.jobs)
	if err != nil {
		logger.E(tag, "parse failed:%+v", err)
		return
	}

	wg := &sync.WaitGroup{}
	for _, item := range jobSelected {
		wg.Add(1)
		go func(job JobHandler) {
			defer wg.Done()
			defer processMark(tag, "job name:"+job.Name)()
			if err := job.Do(); err != nil {
				logger.E(tag, "[%s] run failed,error: %+v", job.Name, err)
			}
		}(item)
	}
	wg.Wait()
}

func (js *JobServer) AddServerStartFunc(fns ...bootstrap.ServerStartFunc) {
	js.funcSetter.AddServerStartFunc(fns...)
}

func (js *JobServer) AddServerStopFunc(fns ...bootstrap.ServerStopFunc) {
	js.funcSetter.AddServerStopFunc(fns...)
}

func processMark(tag string, msg string) func() {
	logger.I(tag, "[start] %s", msg)
	return func() {
		logger.I(tag, "[end] %s", msg)
	}
}
