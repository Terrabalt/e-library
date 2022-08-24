package endpoints

import (
	"errors"
	"ic-rhadi/e_library/database"
	"net/http"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type loginPostRequest struct {
	Username string `json:"username"` // User login ID.
	Password string `json:"password"` // Password to verify.
}

func (l *loginPostRequest) Bind(r *http.Request) error {
	if l.Username == "" || l.Password == "" {
		return errors.New("username or password missing")
	}

	return nil
}

var ErrLoginFailed = errors.New("login failed")

func LoginPostEndpoint(db database.UserAccountInterface, sessionAuth *jwtauth.JWTAuth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := r.Context()

		data := &loginPostRequest{}
		if err := render.Bind(r, data); err != nil {
			log.Debug().Err(err).Str("username", data.Username).Msg("Login attempt malformed")
			render.Render(w, r, BadRequestError(err))
			return
		}

		session, err := db.Login(ctx, data.Username, data.Password, false)
		if err != nil {
			log.Debug().Err(err).Str("username", data.Username).Msg("Login attempt failed")
			if err == database.ErrAccountNotActive {
				render.Render(w, r, ValidationFailedError(err))
			} else {
				render.Render(w, r, ValidationFailedError(ErrLoginFailed))
			}
			return
		}

		sendNewToken(sessionAuth, tokenClaimsSchema{data.Username, session}, w, r)
	}
}
