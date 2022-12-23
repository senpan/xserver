package bootstrap

import (
	"github.com/senpan/xserver/logger"
)

type ServerStartFunc func() error

func SetXServerLogger(log logger.Logger) ServerStartFunc {
	return func() error {
		logger.SetLogger(log)
		return nil
	}
}
