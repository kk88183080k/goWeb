package binding

import "net/http"

type Binding interface {
	Name() string
	Bind(r *http.Request, v any) error
}

var JsonBind jsonBinding = jsonBinding{}
var XmlBind xmlBinding = xmlBinding{}
