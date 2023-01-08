package render

import (
	"html/template"
	"net/http"
)

type Render interface {
	Render(w http.ResponseWriter) error
	WriteContentType(w http.ResponseWriter)
}

func WriteContentTypeValue(w http.ResponseWriter, contextTypeVal string) {
	w.Header().Add("Content-Type", contextTypeVal)
}

type HTMLRender struct {
	Template *template.Template
}
