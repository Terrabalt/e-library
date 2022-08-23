package endpoints

import (
	"context"
	"errors"
	"ic-rhadi/e_library/database"
	"net/http"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
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

type googleClaimsSchema struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FullName      string `json:"name"`
	AccountId     string `json:"sub"`
}

func validateGoogleToken(ctx context.Context, token string) (*googleClaimsSchema, error) {
	return nil, errors.New("endpoint not implemented")
}

func LoginGoogleEndpoint(db database.DB, sessionAuth *jwtauth.JWTAuth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := r.Context()
		data := &loginGoogleRequest{}
		if err := render.Bind(r, data); err != nil {
			log.Debug().Err(err).Msg("Google token body malformed")
			render.Render(w, r, BadRequestError(err))
			return
		}

		sch, err := validateGoogleToken(ctx, data.GoogleToken)
		if err != nil {
			log.Debug().Err(err).Str("username", data.GoogleToken).Msg("Google token validation failed")
			render.Render(w, r, ValidationFailedError(err))
			return
		}

		session, err := db.Login(ctx, sch.Email, sch.AccountId, true)
		if err != nil {
			log.Debug().Err(err).Str("username", sch.Email).Msg("Login attempt failed")
			if err == database.ErrAccountNotActive {
				render.Render(w, r, ValidationFailedError(err))
			} else {
				render.Render(w, r, ValidationFailedError(ErrLoginFailed))
			}
			return
		}

		sendNewToken(sessionAuth, tokenClaimsSchema{sch.Email, session}, w, r)
	}
}
