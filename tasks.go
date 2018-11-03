package nyn

import taskqueue "github.com/touee/nyn/task-queue"

// Task 是爬虫任务的接口
type Task interface {
	Fetch(c *Crawler, self Task) (payload interface{}, err error)
	Process(c *Crawler, self Task, payload interface{}) (result taskqueue.ProcessResult, err error)
}

// WithFilter 实现此接口的任务会被调用 filter 以进行过滤
type WithFilter interface {
	Filter(c *Crawler, self Task) (FilterResult, error)
}

/*
// WithTaskModifier 实现此接口的任务在 fetch 前, 经由 GetModifiedTask 获得修改后的任务
type WithTaskModifier interface {
	GetModifiedTask(c *Crawler) (newTask Task, err error)
}
*/

// WithPayloadDecorator 对 fetch 返回的 payload 进行装饰
type WithPayloadDecorator interface {
	DecoratePayload(c *Crawler, self Task, payload interface{}) (decoratedPayload interface{}, err error)
}
