package render

import (
	"errors"
	"net/http"
	"strconv"
)

type Redirect struct {
	Url     string
	Status  int
	Request *http.Request
}

func (r *Redirect) Render(w http.ResponseWriter) error {
	if (r.Status < http.StatusMultipleChoices || r.Status > http.StatusPermanentRedirect) && r.Status != http.StatusCreated {
		return errors.New("Redirect status error, status:" + strconv.Itoa(r.Status))
	}
	http.Redirect(w, r.Request, r.Url, r.Status)
	return nil
}

func (r *Redirect) WriteContentType(w http.ResponseWriter) {
}
