package msgo

import (
	"errors"
	"fmt"
	"github.com/kk88183080k/goWeb/msgo/mserror"
	"net/http"
	"runtime"
	"strings"
)

func Recovery(fn Handler) Handler {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {

				// 判断是否是自定义的错误
				if e := err.(error); e != nil {
					var msErr *mserror.MsError
					if errors.As(e, &msErr) {
						msErr.ExecuteResult()
						return
					}
				}

				ctx.Logger.Error(detailMsg(err))
				ctx.Fail(http.StatusInternalServerError, "服务器内部错误")
			}
		}()
		fn(ctx)
	}
}

func detailMsg(err any) string {
	var sb strings.Builder
	var pcs = make([]uintptr, 32)
	n := runtime.Callers(3, pcs)
	sb.WriteString(fmt.Sprintf("%v\n", err))
	for _, pc := range pcs[:n] {
		//函数
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		sb.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return sb.String()
}
