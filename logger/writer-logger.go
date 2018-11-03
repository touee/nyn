package logger

import (
	"io"
	"sync"
	"time"
)

// WriterLogger 是将日志输出到 writer 的 logger
type WriterLogger struct {
	formatter Formatter
	writer    io.Writer
	lock      sync.Mutex

	level Level
}

// NewWriterLogger 返回一个新的 WriterLogger
func NewWriterLogger(writer io.Writer, level Level) *WriterLogger {
	return &WriterLogger{
		formatter: NewSimpleFormatter(),
		writer:    writer,

		level: level,
	}
}

// Log 将日志写入 writer
func (m *WriterLogger) Log(module string,
	time time.Time, level Level,
	message string, values Fields) {
	if level <= m.level {
		var log = m.formatter.Format(module, time, level, message, values)
		m.lock.Lock()
		m.writer.Write([]byte(log + "\n"))
		m.lock.Unlock()
	}
}
