package render

import (
	"encoding/xml"
	"net/http"
)

type Xml struct {
	Data any
}

func (x *Xml) Render(w http.ResponseWriter) error {
	x.WriteContentType(w)
	return xml.NewEncoder(w).Encode(x.Data)
}

func (x *Xml) WriteContentType(w http.ResponseWriter) {
	WriteContentTypeValue(w, "application/xml; charset=utf-8")
}
