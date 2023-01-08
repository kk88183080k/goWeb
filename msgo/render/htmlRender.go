package render

import (
	"html/template"
	"net/http"
)

type HtmlOptionsRender struct {
	Name       string
	Data       any
	Template   *template.Template
	IsTemplate bool
}

func (h *HtmlOptionsRender) Render(w http.ResponseWriter) error {
	h.WriteContentType(w)

	if !h.IsTemplate {
		_, err := w.Write([]byte(h.Data.(string)))
		return err
	}

	return h.Template.ExecuteTemplate(w, h.Name, h.Data)
}

func (h *HtmlOptionsRender) WriteContentType(w http.ResponseWriter) {
	WriteContentTypeValue(w, "text/html; charset=utf-8")
}
