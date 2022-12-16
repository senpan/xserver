package bootstrap

type FuncSetter struct {
	beforeFunc []ServerStartFunc
	afterFunc  []ServerStopFunc
}

func NewFuncSetter() *FuncSetter {
	return &FuncSetter{}
}

func (fs *FuncSetter) AddServerStartFunc(fns ...ServerStartFunc) {
	for _, fn := range fns {
		fs.beforeFunc = append(fs.beforeFunc, fn)
	}
}

func (fs *FuncSetter) AddServerStopFunc(fns ...ServerStopFunc) {
	for _, fn := range fns {
		fs.afterFunc = append(fs.afterFunc, fn)
	}
}

func (fs *FuncSetter) RunServerStartFunc() error {
	for _, fn := range fs.beforeFunc {
		err := fn()
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *FuncSetter) RunServerStopFunc() {
	for _, fn := range fs.afterFunc {
		fn()
	}
}
