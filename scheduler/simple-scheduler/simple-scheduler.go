package sscheduler

import (
	"path"

	"github.com/touee/nyn/scheduler"
	"github.com/touee/nyn/task-queue/sqlite3"

	"github.com/touee/nyn/logger"

	datapacker "github.com/touee/nyn/data-packer"
	"github.com/touee/nyn/task-queue"

	"github.com/touee/nyn/type-manager"

	workerpool "github.com/touee/nyn/worker-pool"
)

// SimpleScheduler 是一个简单的调度器, 实现了并发执行任务
type SimpleScheduler struct {
	//quitWaitChan chan []error

	workerPool  *workerpool.WorkerPool
	typeManager *typemanager.TypeManager
	taskQueue   taskqueue.TaskQueue

	logger      logger.Logger
	mode        scheduler.Mode
	taskHandler scheduler.TaskHandler
}

// SimpleSchedulerOptions 包含了 SimpleScheduler 的各类设置
type SimpleSchedulerOptions struct {
	// Dir 是其工作目录
	Dir string

	// DefaultWokerPoolWorkers 是并发中的 worker 的数量
	DefaultWorkerPoolWorkers int

	// TypeManagerFileName 是的文件名
	TypeManagerFileName string
	// DataPacker 是 TaskTypeManager 所用的 packer
	// 缺省为 ObjpackPacker
	DataPacker datapacker.Packer

	// TaskQueue 是 SimpleScheduler 使用的 TaskQueue
	// 缺省为 "./worker-pool".WorkerPool
	TaskQueue taskqueue.TaskQueue
}

// NewSimpleScheduler 返回一个新的简单调度器
func NewSimpleScheduler(opts SimpleSchedulerOptions) (s *SimpleScheduler, err error) {
	s = new(SimpleScheduler)

	// 初始化 worker pool
	if opts.DefaultWorkerPoolWorkers == 0 {
		opts.DefaultWorkerPoolWorkers = 5
	}
	s.workerPool = workerpool.NewWorkerPool(opts.DefaultWorkerPoolWorkers)

	// 初始化类型 manager (以及 data packer)
	if opts.TypeManagerFileName == "" {
		opts.TypeManagerFileName = "task-type.s3db"
	}
	opts.TypeManagerFileName = path.Join(opts.Dir, opts.TypeManagerFileName)

	if opts.DataPacker == nil {
		opts.DataPacker = datapacker.DefaultPacker
	}

	if s.typeManager, err = typemanager.OpenTypeManager(opts.TypeManagerFileName, opts.DataPacker); err != nil {
		return nil, err
	}

	// 初始化 任务队列
	if opts.TaskQueue == nil {
		var queuePath = path.Join(opts.Dir, "queue.s3db")
		if s.taskQueue, err = s3queue.Open(queuePath); err != nil {
			return nil, err
		}
	} else {
		s.taskQueue = opts.TaskQueue
	}

	//s.quitWaitChan = make(chan []error)

	return s, nil
}

// SetTaskHandler 设置调度器的 TaskHandler
func (s *SimpleScheduler) SetTaskHandler(handler scheduler.TaskHandler) {
	s.taskHandler = handler
}

// SetMode 设置 mode
func (s *SimpleScheduler) SetMode(mode scheduler.Mode) {
	s.mode = mode
}
