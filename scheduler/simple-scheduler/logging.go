package sscheduler

import (
	"time"

	"github.com/touee/nyn/logger"
)

// SetLogger 设置 logger
func (s *SimpleScheduler) SetLogger(logger logger.Logger) {
	s.logger = logger
	s.workerPool.SetLogger(s.logger)
}

func (s *SimpleScheduler) log(level logger.Level, message string, values logger.Fields) {
	if s.logger != nil {
		s.logger.Log("SimpleSchedluer", time.Now(), level, message, values)
	}
}
