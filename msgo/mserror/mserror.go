package mserror

type MsError struct {
	err error
	fn  ErrorFn
}

func (e *MsError) Error() string {
	return e.err.Error()
}

func Default() *MsError {
	return &MsError{}
}

func (e *MsError) Put(err error) {
	e.checkErr(err)
}

func (e *MsError) checkErr(err error) {
	e.err = err
	if e.err != nil {
		panic(e)
	}
}

// 自定义错误处理函数
type ErrorFn func(e *MsError)

func (e *MsError) Result(errFn ErrorFn) {
	e.fn = errFn
}

func (e *MsError) ExecuteResult() {
	e.fn(e)
}
