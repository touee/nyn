package workerpool

import (
	"time"

	"github.com/touee/nyn/logger"
)

// SetLogger 设置 logger
func (pool *WorkerPool) SetLogger(logger logger.Logger) {
	pool.logger = logger
}

func (pool *WorkerPool) log(level logger.Level, message string, values logger.Fields) {
	if pool.logger != nil {
		pool.logger.Log("WorkerPool", time.Now(), level, message, values)
	}
}
