package logger

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Formatter 是日志的格式化函数, 将日志格式化为字符串
// 返回的字符串应没有换行
type Formatter interface {
	Format(module string,
		time time.Time, level Level,
		message string, values Fields) string
}

// SimpleFormatter 是一个简单的日志格式化器
type SimpleFormatter struct {
	lastLogTime time.Time
	timeLock    sync.Mutex
}

// Format 格式化日志
func (f *SimpleFormatter) Format(module string,
	logTime time.Time, level Level,
	message string, values Fields) string {
	var logFullTime bool
	f.timeLock.Lock()
	var now = time.Now()
	if now.Day() != f.lastLogTime.Day() ||
		now.Sub(f.lastLogTime).Hours() >= 24 {
		logFullTime = true
	}
	f.lastLogTime = now
	f.timeLock.Unlock()

	var allValues string
	for _, value := range values {
		allValues += value.Key + "="
		var jsonizedValue, err = json.Marshal(value.Value)
		if err != nil {
			allValues += "ERROR_LOG_FIELD_VALUE(" + err.Error() + ")"
		} else {
			allValues += string(jsonizedValue)
		}
		allValues += " "
	}
	if len(allValues) > 0 && allValues[len(allValues)-1] == ' ' {
		allValues = allValues[:len(allValues)-1]
	}

	var timeFormat = "15:04:05"
	if logFullTime {
		timeFormat = time.RFC3339
	}

	return fmt.Sprintf("[%s]<%s>%s: %s %s", logTime.Format(timeFormat), level.String(), module, message, allValues)
}

// NewSimpleFormatter 返回一个新的 SimpleFormatter
func NewSimpleFormatter() *SimpleFormatter {
	return &SimpleFormatter{}
}
