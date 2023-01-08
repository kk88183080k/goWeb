package msgo

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	greenBg   = "\033[97;42m"
	whiteBg   = "\033[90;47m"
	yellowBg  = "\033[90;43m"
	redBg     = "\033[97;41m"
	blueBg    = "\033[97;44m"
	magentaBg = "\033[97;45m"
	cyanBg    = "\033[97;46m"
	green     = "\033[32m"
	white     = "\033[37m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	reset     = "\033[0m"
)

type loggingConfig struct {
	Formatter LoggerFormatter
	Out       io.Writer
	IsColor   bool
}

// LoggerFormatter 日志的格式化方法
type LoggerFormatter = func(params *LogFormatterParams) string
type LogFormatterParams struct {
	Request        *http.Request
	TimeStamp      time.Time
	StatusCode     int
	Latency        time.Duration
	ClientIP       net.IP
	Method         string
	Path           string
	IsDisplayColor bool
}

func (p *LogFormatterParams) ResetColor() string {
	return reset
}

func (p *LogFormatterParams) StatusCodeColor() string {
	switch p.StatusCode {
	case http.StatusOK:
		return green
	default:
		return red
	}
}

// defaultFormatter 日志的默认的格式化方法
var defaultFormatter = func(params *LogFormatterParams) string {
	if params.Latency > time.Minute {
		params.Latency = params.Latency.Truncate(time.Second)
	}

	if params.IsDisplayColor {
		resetColor := params.ResetColor()
		statusCodeColor := params.StatusCodeColor()
		return fmt.Sprintf("%s [msgo] %s |%s %v %s| %s %3d %s |%s %13v %s| %15s  |%s %-7s %s %s %#v %s",
			yellow, resetColor, blue, params.TimeStamp.Format("2006/01/02 - 15:04:05"), resetColor,
			statusCodeColor, params.StatusCode, resetColor,
			red, params.Latency, resetColor,
			params.ClientIP,
			magenta, params.Method, resetColor,
			cyan, params.Path, resetColor,
		)
	}

	return fmt.Sprintf("[msgo] %v | %3d | %13v | %15s |%-7s %#v",
		params.TimeStamp.Format("2006/01/02 - 15:04:05"),
		params.StatusCode,
		params.Latency, params.ClientIP, params.Method, params.Path,
	)

}

// defaultWrite 日志的默认的输出
var defaultWrite io.Writer = os.Stdout

func loggingWithConfig(config loggingConfig, nextFn Handler) Handler {
	if config.Formatter == nil {
		config.Formatter = defaultFormatter
	}

	if config.Out == nil {
		config.Out = defaultWrite
		config.IsColor = true
	}
	return func(ctx *Context) {
		// 执行前
		r := ctx.R
		path := r.URL.Path
		rawQuery := r.URL.RawQuery
		startTime := time.Now()

		nextFn(ctx)

		// 执行后
		params := &LogFormatterParams{}
		params.Request = r
		params.TimeStamp = time.Now()
		params.Latency = params.TimeStamp.Sub(startTime)
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(ctx.R.RemoteAddr))
		params.ClientIP = net.ParseIP(ip)
		params.Method = r.Method
		params.StatusCode = ctx.StatusCode

		if rawQuery != "" {
			path += rawQuery
		}
		params.Path = path

		params.IsDisplayColor = config.IsColor
		fmt.Fprintln(config.Out, config.Formatter(params))
	}
}

// Logging 日志中间件调用方法
//func Logging(nextFn Handler) Handler {
//	return loggingWithConfig(loggingConfig{}, nextFn)
//}

func Logging(next Handler) Handler {
	return func(ctx *Context) {
		// 执行前
		r := ctx.R
		path := r.URL.Path
		rawQuery := r.URL.RawQuery
		startTime := time.Now()

		next(ctx)

		// 执行后
		params := LogFormatterParams{}
		//params.Request = r
		params.TimeStamp = time.Now()
		params.Latency = params.TimeStamp.Sub(startTime)
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(ctx.R.RemoteAddr))
		params.ClientIP = net.ParseIP(ip)
		params.Method = r.Method
		params.StatusCode = ctx.StatusCode

		if rawQuery != "" {
			path += rawQuery
		}
		params.Path = path
		reqLog, err := json.Marshal(params)
		if err != nil {
			ctx.Logger.Error(err)
			return
		}
		ctx.Logger.Debug(string(reqLog))
	}
}
