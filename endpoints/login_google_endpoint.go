package endpoints

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
)

type loginGoogleRequest struct {
	GoogleToken string `json:"token"` // User login ID.
}

func (l *loginGoogleRequest) Bind(r *http.Request) error {
	if l.GoogleToken == "" {
		return errors.New("token missing")
	}

	return nil
}

func LoginGoogleEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		//ctx := r.Context()
		data := &loginGoogleRequest{}
		if err := render.Bind(r, data); err != nil {
			r.Response.StatusCode = http.StatusBadRequest
			return
		}
	}
}
