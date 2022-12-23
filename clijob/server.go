package clijob

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/senpan/xserver/bootstrap"
	"github.com/senpan/xserver/logger"
)

type JobServer struct {
	funcSetter *bootstrap.FuncSetter
	opts       Options
	jobs       map[string]JobHandler
	cmdParser  CmdParser
	exit       chan struct{} // 退出
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

	if err = js.funcSetter.RunStartFunc(); err != nil {
		return nil
	}

	go js.regSignal()
	go js.doJob()

	<-js.exit
	return nil
}

func (js *JobServer) waitShutdown() {
	logger.GetLogger().Infof("xserver.JobServer", "Stop...")
	js.funcSetter.RunStopFunc()

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
		logger.GetLogger().Errorf(tag, "parse failed:%+v", err)
		return
	}

	wg := &sync.WaitGroup{}
	for _, item := range jobSelected {
		wg.Add(1)
		go func(job JobHandler) {
			defer wg.Done()
			defer processMark(tag, "job name:"+job.Name)()
			if err := job.Do(); err != nil {
				logger.GetLogger().Errorf(tag, "[%s] run failed,error: %+v", job.Name, err)
			}
		}(item)
	}
	wg.Wait()
}

func (js *JobServer) AddStartFunc(fns ...bootstrap.ServerStartFunc) {
	js.funcSetter.AddStartFunc(fns...)
}

func (js *JobServer) AddStopFunc(fns ...bootstrap.ServerStopFunc) {
	js.funcSetter.AddStopFunc(fns...)
}

func processMark(tag string, msg string) func() {
	logger.GetLogger().Infof(tag, "[start] %s", msg)
	return func() {
		logger.GetLogger().Infof(tag, "[end] %s", msg)
	}
}
