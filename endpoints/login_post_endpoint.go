package endpoints

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
)

type loginPostRequest struct {
	Username string `json:"username"` // User login ID.
	Password string `json:"password"` // Password to verify.
}

func (l *loginPostRequest) Bind(r *http.Request) error {
	if l.Username == "" || l.Password == "" {
		return errors.New("bad request")
	}

	return nil
}

func LoginPostEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		//ctx := r.Context()
		data := &loginPostRequest{}
		if err := render.Bind(r, data); err != nil {
			r.Response.StatusCode = http.StatusBadRequest
			return
		}
	}
}
