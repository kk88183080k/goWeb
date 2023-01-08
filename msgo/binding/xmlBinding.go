package binding

import (
	"encoding/xml"
	"io"
	"net/http"
)

type xmlBinding struct {
}

func (x *xmlBinding) Name() string {
	return "xml"
}

func (x *xmlBinding) Bind(r *http.Request, v any) error {
	return decodeXml(r.Body, v)
}

func decodeXml(body io.ReadCloser, v any) error {
	decoder := xml.NewDecoder(body)
	if err := decoder.Decode(v); err != nil {
		return err
	}
	return validate(v)
}
