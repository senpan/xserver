package bootstrap

import (
	"github.com/senpan/xserver/logger"
)

type ServerStopFunc func()

func CloseXServerLogger() ServerStopFunc {
	return func() {
		logger.GetLogger().Close()
	}
}
