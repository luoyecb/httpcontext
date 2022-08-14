package httpcontext

import (
	"log"
)

// ========================================
// Logger
// ========================================

type Logger interface {
	Log(format string, args ...interface{})
	LogError(msg string, err error)
}

type defaultLogger struct{}

func (l *defaultLogger) Log(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (l *defaultLogger) LogError(msg string, err error) {
	log.Printf("[httpcontext] %s, error:%v\n", msg, err)
}
