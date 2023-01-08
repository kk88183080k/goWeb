package render

import (
	"fmt"
	"github.com/kk88183080k/goWeb/msgo/utils"
	"net/http"
)

type String struct {
	Format string
	Data   []any
}

func (s *String) Render(w http.ResponseWriter) error {
	s.WriteContentType(w)
	if len(s.Data) > 0 {
		_, err := fmt.Fprintf(w, s.Format, s.Data...)
		return err
	}

	_, err := w.Write(utils.StringToBytes(s.Format))
	return err
}

func (s *String) WriteContentType(w http.ResponseWriter) {
	WriteContentTypeValue(w, "text/plain; charset=utf-8")
}
