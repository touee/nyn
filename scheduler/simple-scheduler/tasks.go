package sscheduler

// RegisterTaskType 注册一个任务类型
func (s *SimpleScheduler) RegisterTaskType(dummyTask interface{}) (err error) {
	return s.typeManager.Register(dummyTask)
}
