package logs

import (
	"encoding/json"
	"fmt"
	"time"
)

type JsonFormatter struct {
}

func (f *JsonFormatter) Formatter(para *LoggerFormatPara) string {

	msg := make(map[string]any)
	msg["time"] = time.Now()
	msg["level"] = para.Level.Level()
	msg["msg"] = para.Msg
	msg["fields"] = para.Fields

	jsonStr, _ := json.Marshal(msg)
	return fmt.Sprintf("[msgo] %s \n", jsonStr)
}
