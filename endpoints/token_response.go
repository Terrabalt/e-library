package endpoints

import (
	"net/http"

	"github.com/go-chi/render"
)

type tokenResponse struct {
	Session   string `json:"session_token"`
	Refresh   string `json:"refresh_token"`
	Scheme    string `json:"scheme"`
	ExpiresAt string `json:"expires_at"`
}

func (token tokenResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusOK)
	w.Header().Set("content-type", "application/json")
	return nil
}
