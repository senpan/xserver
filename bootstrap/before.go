package bootstrap

import (
	logger "github.com/senpan/xlogger"
)

type ServerStartFunc func() error

func InitLogger(version string) ServerStartFunc {
	return func() error {
		logger.InitXLogger(version)
		return nil
	}
}
