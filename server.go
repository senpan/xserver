package xserver

import "github.com/senpan/xserver/bootstrap"

// XServer Server接口
type XServer interface {
	AddStartFunc(fns ...bootstrap.ServerStartFunc)
	AddStopFunc(fns ...bootstrap.ServerStopFunc)
	Serve() error
}
