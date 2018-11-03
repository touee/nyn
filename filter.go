package nyn

import taskqueue "github.com/touee/nyn/task-queue"

// FilterResult 是 filter 返回的结果
type FilterResult int

const (
	// FilterResultPass 同 ProcessResultSuccessful
	FilterResultPass FilterResult = FilterResult(taskqueue.ProcessResultSuccessful)
	// FilterResultShouldBeExcluded 同 ProcessResultShouldBeExcluded
	FilterResultShouldBeExcluded FilterResult = FilterResult(taskqueue.ProcessResultShouldBeExcluded)
	// FilterResultShouldBeFrozen 同 ProcessResultShouldBeFrozen
	FilterResultShouldBeFrozen FilterResult = FilterResult(taskqueue.ProcessResultShouldBeFrozen)
)
