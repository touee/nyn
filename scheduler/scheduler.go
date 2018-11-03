package scheduler

import (
	"github.com/touee/nyn/logger"
	taskqueue "github.com/touee/nyn/task-queue"
)

// Scheduler 代表爬虫的调度器
type Scheduler interface {
	SetTaskHandler(handler TaskHandler)
	// TODO: SetGivenUpHandler
	SetLogger(logger logger.Logger)
	SetMode(mode Mode)

	RegisterTaskType(dummyTask interface{}) (err error)

	AddTask(task interface{}, opts AddTaskOptions) (err error)

	Run()
	WaitQuit() (errs []error)
}

// TaskHandler 是任务的 handler
type TaskHandler func(interface{}) taskqueue.ProcessResult

// AddTaskOptions = taskqueue.EnqueueOptions
type AddTaskOptions = taskqueue.EnqueueOptions

// Mode 代表调度器的模式
type Mode int

const (
	// ModeCollector 爬虫将会在达成目标 (没有更多任务) 时退出
	ModeCollector Mode = iota
	// ModeMonitor 爬虫在没有任务后会持续等待
	ModeMonitor
)
