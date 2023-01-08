package msgo

type RError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type R struct {
	RError
	Data any `json:"data"`
}

func (r RError) Error() string {
	return r.Msg
}

func DefaultR() *R {
	return &R{}
}

func (r *R) Success(code int, msg string, data any) *R {
	r.setValues(code, msg, data)
	return r
}

func (r *R) Fail(code int, msg string) *R {
	r.setValues(code, msg, nil)
	return r
}

func (r *R) setValues(code int, msg string, data any) {
	r.Code = code
	r.Msg = msg
	r.Data = data
}

func (r *R) Response() any {
	if r.Data == nil {
		return &RError{Code: r.Code, Msg: r.Msg}
	}
	return r
}
