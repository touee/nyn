package taskqueue

import "errors"

var (
	// ErrAttemptToChangeStatusOfTaskInProcessing 是当在 OverwritingFor 中设置了 TaskStatusProcessing 时返回的错误
	ErrAttemptToChangeStatusOfTaskInProcessing = errors.New("taskqueue: Change status of tasks in prcoessing is not allowed")

	// ErrAttemptToChangeStatusOfTaskToProcessing 是当尝试将任务状态手动设置为处理中时返回的错误
	ErrAttemptToChangeStatusOfTaskToProcessing = errors.New("taskqueue: Change status of tasks to prcoessing is not allowed")

	// ErrInvaildPriority 是当优先级非正数时返回的错误
	ErrInvaildPriority = errors.New("taskqueue: Priority must > 0")

	// ErrTaskExists 是当任务本身就存在(不一定在等待队列中)时返回的错误
	ErrTaskExists = errors.New("taskqueue: Task already exists")

	// ErrNoTasksInPending 是没有可返回的等待中的任务时返回的错误
	ErrNoTasksInPending = errors.New(`taskqueue: No task in the pending queue`)
)
