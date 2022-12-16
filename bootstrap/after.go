package bootstrap

import (
	logger "github.com/senpan/xlogger"
)

type ServerStopFunc func()

func CloseLogger() ServerStopFunc {
	return func() {
		logger.Close()
	}
}
