package bootstrap

type FuncSetter struct {
	befFns []ServerStartFunc
	aftFns []ServerStopFunc
}

func NewFuncSetter() *FuncSetter {
	return &FuncSetter{}
}

func (fs *FuncSetter) AddStartFunc(fns ...ServerStartFunc) {
	fs.befFns = append(fs.befFns, fns...)
}

func (fs *FuncSetter) AddStopFunc(fns ...ServerStopFunc) {
	fs.aftFns = append(fs.aftFns, fns...)
}

func (fs *FuncSetter) RunStartFunc() error {
	for _, fn := range fs.befFns {
		err := fn()
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *FuncSetter) RunStopFunc() {
	for _, fn := range fs.aftFns {
		fn()
	}
}
