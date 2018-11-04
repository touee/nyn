package nyn

import (
	"fmt"
	"os"
	"path"

	"github.com/touee/nyn/logger"
	"github.com/touee/nyn/scheduler"
	"github.com/touee/nyn/scheduler/simple-scheduler"
	taskqueue "github.com/touee/nyn/task-queue"
	typemanager "github.com/touee/nyn/type-manager"
)

// Crawler 是爬虫
type Crawler struct {
	defaultLogFile *os.File
	logger         logger.Logger

	scheduler scheduler.Scheduler

	Global map[string]interface{}
	//Shared map[string]map[string]interface{}
}

// CrawlerOptions 包含了爬虫的各类设置选项
type CrawlerOptions struct {
	// Dir 是爬虫的工作目录. 若 dir 已存在则尝试从 dir 复原爬虫的工作状态, 若 dir 不存在则创建之
	Dir string

	// NoDefaultFileLogger 若为真, 则不使用默认的 stdout logger
	NoDefaultStdoutLogger bool
	// DefaultLoggerLevel 是爬虫默认 stdout logger 的级别
	// 如果为 0 则会认作 LInfo
	DefaultStdoutLoggerLogLevel logger.Level
	// NoDefaultFileLogger 若为真, 则不使用默认的 file logger
	NoDefaultFileLogger bool
	// DefaultLoggerLevel 是爬虫默认 file logger 的级别
	// 如果为 0 则会认作 LInfo
	DefaultFileLoggerLogLevel logger.Level
	// Logger logger, 若 RemoveDefaultLogger 为真, 则为唯一的 logger, 否则与默认 logger 共存
	Logger logger.Logger

	// Scheduler 如果不为 nil, 则将替换掉爬虫的默认调度器
	Scheduler scheduler.Scheduler

	// DefaultWokerPoolWorkers 是默认任务池的 worker 数
	DefaultWorkerPoolWorkers int

	Mode scheduler.Mode
}

// NewCrawler 返回一个新的 Crawler
func NewCrawler(opts CrawlerOptions) (c *Crawler, err error) {
	c = new(Crawler)

	{ //< 处理传入的 options
		if opts.NoDefaultStdoutLogger && opts.NoDefaultFileLogger {
			c.logger = opts.Logger
		} else {

			var defaultStdoutLoggerLevel = opts.DefaultStdoutLoggerLogLevel
			var defaultFileLoggerLevel = opts.DefaultFileLoggerLogLevel
			if defaultStdoutLoggerLevel == 0 {
				defaultStdoutLoggerLevel = logger.LInfo
			}
			if defaultFileLoggerLevel == 0 {
				defaultFileLoggerLevel = logger.LInfo
			}

			if opts.Logger != nil {
				c.logger = logger.MultiLogger{opts.Logger}
			} else {
				c.logger = logger.MultiLogger{}
			}

			if !opts.NoDefaultStdoutLogger {
				c.logger = append(c.logger.(logger.MultiLogger),
					logger.NewWriterLogger(os.Stdout, defaultStdoutLoggerLevel))
			}

			if !opts.NoDefaultFileLogger {
				var logFilePath = path.Join(opts.Dir, "logs.log")
				if c.defaultLogFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_RDWR, 0644); err != nil {
					c.log(logger.LError, "unable to open log file.", logger.Fields{{"path", logFilePath}, {"error", err}})
					c.defaultLogFile = nil
				} else {
					c.logger = append(c.logger.(logger.MultiLogger),
						logger.NewWriterLogger(c.defaultLogFile, defaultFileLoggerLevel))
				}
			}

		}

		c.log(logger.LDebug, "Hello World!", logger.Fields{{"answer", 42}})

		if opts.Scheduler != nil {
			c.scheduler = opts.Scheduler
		} else {
			c.scheduler, err = sscheduler.NewSimpleScheduler(sscheduler.SimpleSchedulerOptions{
				Dir:                      opts.Dir,
				DefaultWorkerPoolWorkers: opts.DefaultWorkerPoolWorkers,
			})
			if err != nil {
				return nil, err
			}
		}
		c.scheduler.SetMode(opts.Mode)
		c.scheduler.SetLogger(c.logger)

		// TODO: …
	}

	{ //< 处理环境变量
		// TODO: …
	}

	c.Global = make(map[string]interface{})

	c.scheduler.SetTaskHandler(
		func(_task interface{}) taskqueue.ProcessResult {
			var err error
			var task = _task.(Task)

			// 由于以下步骤出现错误时的处理都一样, 因此包装成一个函数
			var handleErr = func(c *Crawler, step string, err error) taskqueue.ProcessResult {
				type MayBeTemporary interface{ Temporary() bool }
				if tErr, ok := err.(MayBeTemporary); ok && tErr.Temporary() { //< 遇到了临时的网络错误, 以后再重试
					c.log(logger.LWarning, fmt.Sprintf("%s returns a temporary error. will retry later.", step), logger.Fields{
						{"task-type", typemanager.GetTypeName(task)},
						{"task", task},
						{"error", err.Error()},
					})
					return taskqueue.ProcessResultRetry
				} else if err != nil { //< 遇到了通常的错误, 通知 scheduler 任务失败
					c.log(logger.LError, fmt.Sprintf("%s returns a error. will retry later if hasn't reached attempt limit.", step), logger.Fields{
						{"task-type", typemanager.GetTypeName(task)},
						{"task", task},
						{"error", err.Error()},
					})
					return taskqueue.ProcessResultFailed
				}
				panic("?")
			}

			// 1. filter
			var filterResult FilterResult
			if withFilter, ok := task.(WithFilter); ok { //< 任务类型实现了 filter, 因此在 fetch 前先进行筛选
				filterResult, err = withFilter.Filter(c, task)
				if err != nil {
					return handleErr(c, "task.Filter()", err)
				}
				switch filterResult { //< 筛选结果
				case FilterResultShouldBeExcluded: //< 该任务应该被除外
					c.log(logger.LTrace, "Due to task.Filter()'s decision, the task have been excluded", logger.Fields{
						{"task-type", typemanager.GetTypeName(task)},
						{"task-data", task},
					})
					return taskqueue.ProcessResultShouldBeExcluded
				case FilterResultShouldBeFrozen: //< 该任务应该被冻结
					c.log(logger.LTrace, "Due to task.Filter()'s decision, the task have been frozen", logger.Fields{
						{"task-type", typemanager.GetTypeName(task)},
						{"task-data", task},
					})
					return taskqueue.ProcessResultShouldBeFrozen
				case FilterResultPass: //< 一切正常
				default: //< ?
					panic("?")
				}
			}

			// 2. fetch
			var payload interface{}
			payload, err = task.Fetch(c, task)
			if err != nil {
				return handleErr(c, "task.Fetch()", err)
			}

			// 3. decorate payload
			if withPayloadDecorator, ok := task.(WithPayloadDecorator); ok {
				payload, err = withPayloadDecorator.DecoratePayload(c, task, payload)
				if err != nil {
					return handleErr(c, "task.Fetch()", err)
				}
			}

			// 4. process
			var processResult taskqueue.ProcessResult
			processResult, err = task.Process(c, task, payload)
			if err != nil {
				return handleErr(c, "task.Fetch()", err)
			}
			return processResult
		},
	)

	// TODO: finish

	return c, nil
}

// RegisterTaskType 注册一个任务类型
func (c *Crawler) RegisterTaskType(dummyTask Task) (err error) {
	return c.scheduler.RegisterTaskType(dummyTask)
}

// RegisterTaskTypes 注册任务类型
func (c *Crawler) RegisterTaskTypes(dummyTasks ...Task) (err error) {
	for _, dummyTask := range dummyTasks {
		if err = c.scheduler.RegisterTaskType(dummyTask); err != nil {
			return err
		}
	}
	return nil
}

// RequestOptions 是任务请求的设置
type RequestOptions = taskqueue.EnqueueOptions

// RequestWithOptions 添加一个任务请求, 带有设置
func (c *Crawler) RequestWithOptions(task Task, opts RequestOptions) (err error) {
	err = c.scheduler.AddTask(task, opts)
	if _, ok := err.(typemanager.TypeNotRegisteredError); ok {
		panic(err)
	}
	return err
}

// Request 添加一个任务请求
func (c *Crawler) Request(task Task) (err error) {
	return c.RequestWithOptions(task, RequestOptions{Priority: 1})
}

// Run 启动爬虫
func (c *Crawler) Run() {
	c.scheduler.Run()
}

// WaitQuit 等待爬虫退出
func (c *Crawler) WaitQuit() (errs []error) {
	var err error
	errs = c.scheduler.WaitQuit()
	if c.defaultLogFile != nil {
		if err = c.defaultLogFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
