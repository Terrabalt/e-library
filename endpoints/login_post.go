package endpoints

import (
	"errors"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/sessiontoken"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type loginPostRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (l *loginPostRequest) Bind(r *http.Request) error {
	if l.Email == "" || l.Password == "" {
		return ErrLoginPostMalformed
	}

	return nil
}

var ErrLoginAccountNotActive = errors.New("account has not been activated yet")
var ErrLoginFailed = errors.New("login failed")
var ErrLoginPostMalformed = errors.New("username or password missing")

func LoginPostEndpoint(db database.UserAccountInterface, sessionAuth *jwtauth.JWTAuth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := r.Context()

		data := &loginPostRequest{}
		if err := render.Bind(r, data); err != nil {
			log.Debug().Err(err).Str("username", data.Email).Msg("Login attempt malformed")
			render.Render(w, r, BadRequestError(ErrLoginPostMalformed))
			return
		}

		session, err := db.Login(ctx, data.Email, data.Password)
		if err != nil {
			switch err {
			case database.ErrAccountNotActive:
				render.Render(w, r, UnauthorizedRequestError(ErrLoginAccountNotActive))
			case database.ErrAccountNotFound, database.ErrWrongPass:
				render.Render(w, r, ValidationFailedError(ErrLoginFailed))
			default:
				log.Debug().Err(err).Str("email", data.Email).Msg("Login attempt failed")
				render.Render(w, r, InternalServerError())
			}
			return
		}

		t, tokenString, err := sessiontoken.CreateNewSessionToken(
			sessionAuth,
			sessiontoken.TokenClaimsSchema{
				Email:   data.Email,
				Session: session,
			},
		)
		if err != nil {
			log.Error().Err(err).Caller().Msg("Error encoding new token")
			render.Render(w, r, InternalServerError())
			return
		}

		token := tokenResponse{
			Token:     tokenString,
			Scheme:    "Bearer",
			ExpiresAt: t.Expiration().Format(time.RFC3339),
		}
		render.Render(w, r, &token)
	}
}
