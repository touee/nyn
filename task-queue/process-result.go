package taskqueue

// ProcessResult 是任务处理的结果
type ProcessResult int

const (
	// ProcessResultSuccessful 代表任务处理成功
	// 任务之后会标记为已完成
	ProcessResultSuccessful ProcessResult = iota
	// ProcessResultFailed 代表任务失败了, 但是可以根据机会重试
	ProcessResultFailed
	// ProcessResultRetry 代表任务应该重试, 不计为失败
	ProcessResultRetry
	// ProcessResultGivenUp 代表放弃任务
	ProcessResultGivenUp
	// ProcessResultShouldBeExcluded 代表移除任务
	ProcessResultShouldBeExcluded
	// ProcessResultShouldBeFrozen 代表任务应被冻结
	ProcessResultShouldBeFrozen
)
