package render

import (
	"encoding/json"
	"net/http"
)

type Json struct {
	Data any
}

func (j *Json) Render(w http.ResponseWriter) error {
	j.WriteContentType(w)
	dataByte, err := json.Marshal(j.Data)
	if err != nil {
		return err
	}
	_, err = w.Write(dataByte)
	return err
}

func (j *Json) WriteContentType(w http.ResponseWriter) {
	WriteContentTypeValue(w, "application/json; charset=utf-8")
}
