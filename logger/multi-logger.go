package logger

import (
	"time"
)

// MultiLogger 组合多个 logger
type MultiLogger []Logger

// Add 为自身添加一个 logger
func (m *MultiLogger) Add(l Logger) {
	*m = append(*m, l)
}

// Log 传递调用拥有的各 logger 的 Log 方法
func (m MultiLogger) Log(module string,
	time time.Time, level Level,
	message string, values Fields) {
	for _, l := range m {
		l.Log(module, time, level, message, values)
	}
}
