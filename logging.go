package nyn

import (
	"time"

	"github.com/touee/nyn/logger"
)

func (c *Crawler) log(level logger.Level, message string, values logger.Fields) {
	if c.logger != nil {
		c.logger.Log("Crawler", time.Now(), level, message, values)
	}
}

// TaskLog 是供任务使用的 log 函数
func (c *Crawler) TaskLog(level logger.Level, message string, values logger.Fields) {
	if c.logger != nil {
		c.logger.Log("Task", time.Now(), level, message, values)
	}
}
