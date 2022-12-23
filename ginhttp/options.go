package ginhttp

import (
	"time"
)

// Options http server options
type Options struct {
	// run mode, options:debug,release
	Mode string `ini:"mode" yaml:"mode"`
	// TCP address to listen on, ":http" if empty
	Addr string `ini:"addr" yaml:"addr"`
	// grace mode
	Grace bool `ini:"grace" yaml:"grace"`

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration `ini:"readTimeout" yaml:"readTimeout"`
	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	WriteTimeout time.Duration `ini:"writeTimeout" yaml:"writeTimeout"`
	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled. If IdleTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, ReadHeaderTimeout is used.
	IdleTimeout time.Duration `ini:"idleTimeout" yaml:"idleTimeout"`
}

type OptionFunc func(*Options)

func DefaultOptions() Options {
	return Options{
		Addr:         ":10086",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}
}

func WithMode(mode string) OptionFunc {
	return func(o *Options) {
		o.Mode = mode
	}
}

func WithAddr(addr string) OptionFunc {
	return func(o *Options) {
		o.Addr = addr
	}
}

func WithGrace(grace bool) OptionFunc {
	return func(o *Options) {
		o.Grace = grace
	}
}

func WithReadTimeout(readTimeout time.Duration) OptionFunc {
	return func(o *Options) {
		o.ReadTimeout = readTimeout
	}
}

func WithWriteTimeout(writeTimeout time.Duration) OptionFunc {
	return func(o *Options) {
		o.WriteTimeout = writeTimeout
	}
}

func WithIdleTimeout(idleTimeout time.Duration) OptionFunc {
	return func(o *Options) {
		o.IdleTimeout = idleTimeout
	}
}

func WithConfig(conf Options) OptionFunc {
	return func(o *Options) {
		*o = conf
	}
}
