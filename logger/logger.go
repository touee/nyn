package logger

import (
	"time"
)

// Logger 代表一个 logger
type Logger interface {
	Log(
		module string, //< 来源模块
		time time.Time, //< 时间
		level Level, //< 日志级别
		message string, //< 消息 (不含参数)
		values Fields, //< 是事件相关的参数
	)
}

// Level 是日志的级别
type Level uint8

const (
	_ Level = iota
	// LFatal Fatal 级别
	LFatal
	// LError Error 级别
	LError
	// LWarning Warning 级别
	LWarning
	// LInfo Info 级别
	LInfo
	// LDebug Debug 级别
	LDebug
	// LTrace Trace 级别
	LTrace
)

func (l Level) String() string {
	switch l {
	case LTrace:
		return "TRACE"
	case LDebug:
		return "DEBUG"
	case LInfo:
		return "INFO"
	case LWarning:
		return "WARNING"
	case LError:
		return "ERROR"
	case LFatal:
		return "FATAL"
	}
	return "?"
}

// Fields 存放日志中的各类参数
type Fields []struct {
	Key   string
	Value interface{}
}
