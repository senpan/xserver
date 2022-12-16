package clijob

import (
	logger "github.com/senpan/xlogger"
	"github.com/senpan/xtools/flagx"
)

type JobHandler struct {
	Name string // 任务名称
	Task TaskFunc
}

func (j *JobHandler) Do() error {
	if j.Task == nil {
		return logger.NewError("task's function not defined")
	}
	return j.Task()
}

type TaskFunc func() error

type CmdParser interface {
	// JobArgParse 解析命令行参数，并选择对应的job任务
	JobArgParse(jobs map[string]JobHandler) (selectedJobs []JobHandler, err error)
}

type defaultCmdParse struct {
}

func (p *defaultCmdParse) JobArgParse(jobs map[string]JobHandler) (selectedJobs []JobHandler, err error) {
	taskArg := *flagx.GetTask()
	if taskArg == "" {
		return nil, logger.NewError("请使用参数 -task 选择任务, 如：-task testJob")
	}

	job, ok := jobs[taskArg]
	if !ok {
		return nil, logger.NewError("not found [ " + taskArg + " ] job")
	}

	selectedJobs = make([]JobHandler, 0, 1)
	selectedJobs = append(selectedJobs, job)

	return selectedJobs, nil
}
