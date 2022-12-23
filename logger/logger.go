package logger

import "sync"

var (
	logger  Logger
	logLock sync.RWMutex
)

// Logger is the interface for XServer Logger types
type Logger interface {
	Info(tag string, args ...interface{})
	Warn(tag string, args ...interface{})
	Error(tag string, args ...interface{})
	Debug(tag string, args ...interface{})
	Fatal(tag string, args ...interface{})

	Infof(tag string, format string, args ...interface{})
	Warnf(tag string, format string, args ...interface{})
	Errorf(tag string, format string, args ...interface{})
	Debugf(tag string, format string, args ...interface{})
	Fatalf(tag string, format string, args ...interface{})

	Close()
}

func init() {
	SetLogger(new(XServerLogger))
}

// SetLogger sets logger for sdk
func SetLogger(log Logger) {
	logLock.Lock()
	defer logLock.Unlock()
	logger = log
}

func GetLogger() Logger {
	logLock.RLock()
	defer logLock.RUnlock()
	return logger
}
