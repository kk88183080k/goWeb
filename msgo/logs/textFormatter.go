package logs

import (
	"fmt"
	"strings"
	"time"
)

type TextFormatter struct {
}

func (f *TextFormatter) Formatter(para *LoggerFormatPara) string {

	fieldsStr := strings.Builder{}
	if para.Fields != nil {
		count := 0
		for k, v := range para.Fields {
			fmt.Fprintf(&fieldsStr, "%s=%v", k, v)
			if count < len(para.Fields)-1 {
				fmt.Fprintf(&fieldsStr, ",")
				count++
			}
		}
	}

	if para.Color {
		levelColor := f.LevelColor(para.Level)
		msgColor := f.MsgColor(para.Level)
		return fmt.Sprintf("%s [msgo] %s %s%v%s | level= %s %s %s | msg=%s %v %s | fields: %#v \n",
			yellow, reset, blue, time.Now().Format("2006/01/02 - 15:04:05"), reset,
			levelColor, para.Level.Level(), reset, msgColor, para.Msg, reset, fieldsStr.String(),
		)
	}

	return fmt.Sprintf("[msgo] %v | level=%s | msg=%v  | fields: %#v \n",
		time.Now().Format("2006/01/02 - 15:04:05"),
		para.Level.Level(), para.Msg, fieldsStr.String(),
	)
}

func (f *TextFormatter) LevelColor(level LoggerLevel) string {
	switch level {
	case LevelDebug:
		return blue
	case LevelInfo:
		return green
	case LevelError:
		return red
	default:
		return cyan
	}
}

func (f *TextFormatter) MsgColor(level LoggerLevel) string {
	switch level {
	case LevelDebug:
		return blue
	case LevelInfo:
		return green
	case LevelError:
		return red
	default:
		return cyan
	}
}
