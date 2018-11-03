package taskqueue

// TaskQueue 是会由爬虫使用的队列的接口
// 队列插入的规则:
// * 不存在的任务直接插入
// * 存在且等待/冻结的任务, 会修改优先级相关状态及, 并依新旧尝试次数的差值修改相关状态
type TaskQueue interface {
	//Enqueue(prefix, data []byte) (id int, err error)
	EnqueueWithOptions(taskType string, taskData []byte, opts EnqueueOptions) (id int, err error)

	//BatchUnfreezing(filter func(prefix, data []byte) bool) (err error)
	BatchRefreshStatuses(appliedStatuses TaskStatusSet, appliedTaskTypes []string, filter Refresher) (err error)

	DequeueForProcess() (id int, taskType string, taskData []byte, err error)
	ReportProcessResult(id int, result ProcessResult) (err error)

	Close() (err error)
}

// Refresher 用于刷新任务状态
type Refresher func(status TaskStatus, id int, taskType string, taskData []byte) TaskStatus

// EnqueueOptions 是插入任务时的选项
type EnqueueOptions struct {
	// Priority 是优先级
	Priority int
	// ToHead 代表是否插入到队列相应优先级的最前方
	ToHead bool

	// Frozen 如果为真, 任务插入时即为冻结状态
	Frozen bool

	// MaxAttempts 是最多尝试的次数, 如果为 0 则被设置为默认值, 如果为负数则不限
	MaxAttempts int

	// OverwritesFor 是如果已存在要放入的任务, 且该任务的状态在所给状态集内, 则视情况更新:
	// * 对于 TaskStatusPending, 如果优先级更高, 或本身并未被排到队列头部, 则更新优先级;
	// * 对于 TaskStatusProcessing, 如果设置, 会返回 ErrAttemptToOverwriteProcessingTask;
	// * 对于 TaskStatusFrozen 和 TaskStatusGivenUp, 根据此次放入队列的参数重置任务;
	// * 若任务已完成, 对于 TaskStatusFinished, 未设置时会返回 ErrTaskFinished, 设置时会强制重置任务.
	// 其他已存在要放入任务的情况, 会返回特有的错误
	//OverwritesFor ItemStatusSet
}
