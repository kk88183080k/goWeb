package logs

import (
	"fmt"
	"io"
	"os"
	"path"
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

const (
	//100 << 20 100m
	//1 << 20 1m
	//5 << 10 5k
	max_log_size = 100 << 20
)

type LoggerLevel int

const (
	LevelDebug LoggerLevel = iota
	LevelInfo
	LevelError
)

func (l LoggerLevel) Level() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO "
	case LevelError:
		return "ERROR"
	default:
		return ""
	}
}

type log interface {
	Info(v any)
	Debug(v any)
	Error(v any)
}

type Logger struct {
	Level       LoggerLevel
	Outs        []*LogWriter
	Format      LoggerFormat
	Fields      LoggerField
	logPath     string
	LogFileSize int64
}

type LogWriter struct {
	Level LoggerLevel
	Out   io.Writer
}

type LoggerFormat interface {
	Formatter(para *LoggerFormatPara) string
}

type LoggerFormatPara struct {
	Color  bool
	Level  LoggerLevel
	Msg    any
	Fields LoggerField
}

type LoggerField map[string]any

func New() *Logger {
	return &Logger{}
}

func Default() *Logger {
	logger := New()
	logger.Level = LevelDebug
	writer := &LogWriter{
		Level: logger.Level,
		Out:   os.Stdout,
	}
	logger.Outs = append(logger.Outs, writer)
	logger.LogFileSize = max_log_size
	logger.Format = &TextFormatter{}
	return logger
}

func (l *Logger) Info(v any) {
	l.print(LevelInfo, v)
}

func (l *Logger) Debug(v any) {
	l.print(LevelDebug, v)
}

func (l *Logger) Error(v any) {
	l.print(LevelError, v)
}

func (l *Logger) WithField(fields LoggerField) *Logger {
	l.Fields = fields
	return l
}

func (l *Logger) WithFormat(loggerFormat LoggerFormat) *Logger {
	l.Format = loggerFormat
	return l
}

func (l *Logger) SetPath(dir string) *Logger {
	l.logPath = dir

	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err := os.Mkdir(l.logPath, 0644)
		if err != nil {
			fmt.Println("logDir is not exsits, logDir:" + dir)
			panic(err)
		}
	}

	// 创建全部日志文件
	file, err := os.OpenFile(path.Join(l.logPath, "all.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	l.Outs = append(l.Outs, &LogWriter{Level: -1, Out: file})

	// 按级别创建日志文件
	for i := int(l.Level); i <= int(LevelError); i++ {
		file, err := os.OpenFile(path.Join(l.logPath, strings.ToLower(LoggerLevel(i).Level())+".log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			panic(err)
		}
		l.Outs = append(l.Outs, &LogWriter{Out: file, Level: LoggerLevel(i)})
	}

	return l
}

func (l *Logger) print(level LoggerLevel, v any) {
	if l.Level > level {
		// 默认配置的日志级别，高于当前要打印的日志级别；则不打印
		return
	}

	formatPara := &LoggerFormatPara{}
	formatPara.Level = level
	formatPara.Msg = v
	formatPara.Fields = l.Fields

	formatMsg := l.Format.Formatter(formatPara)
	for _, out := range l.Outs {
		if out.Out == os.Stdout {
			formatPara.Color = true
			fmt.Fprintf(out.Out, l.Format.Formatter(formatPara))
			continue
		}
		// 全部日志
		if out.Level == -1 || out.Level == level {
			fmt.Fprintf(out.Out, formatMsg)
			l.checkLogFileSize(out)
			//l.CheckFileSize(out)
		}
	}

}

func (l *Logger) Close() {
	for _, v := range l.Outs {
		if v.Out == os.Stdout {
			continue
		}
		file := v.Out.(*os.File)
		if file != nil {
			file.Close()
		}
	}
}

func (l *Logger) checkLogFileSize(out *LogWriter) {
	file := out.Out.(*os.File)
	if file != nil {
		stat, err := file.Stat()
		if err != nil {
			fmt.Println("日志文件统计大小出错", err)
			return
		}
		if stat.Size() > max_log_size {
			oldFileName := file.Name()
			newFileName := ""
			index := strings.IndexByte(oldFileName, '.')
			if index > 0 {
				newFileName = oldFileName[:index] + "-" + time.Now().Format("2006-01-02-150405") + oldFileName[index:]
			} else {
				newFileName = path.Join(l.logPath, strings.ToLower(out.Level.Level())+"-"+time.Now().Format("2006-01-02-150405")+".log")
			}
			err := file.Close()
			if err != nil {
				fmt.Println("日志文件关闭出错", err)
				return
			}
			err = os.Rename(oldFileName, newFileName)
			if err != nil {
				fmt.Println("日志文件重命名出错", err)
				//return
			}

			newFile, err := os.OpenFile(oldFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				fmt.Println("日志文件创建新文件出错", err)
				return
			}
			out.Out = newFile
		}
	}
}
