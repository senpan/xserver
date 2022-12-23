package consumer

type Options struct {
	path string // mq config path
}

type OptionFunc func(*Options)

func DefaultOptions() Options {
	return Options{
		path: "",
	}
}

func WithPath(path string) OptionFunc {
	return func(o *Options) {
		o.path = path
	}
}
