package endpoints

import (
	"errors"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/googlehelper"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type loginGoogleRequest struct {
	GoogleToken string `json:"token"`
}

func (l *loginGoogleRequest) Bind(r *http.Request) error {
	if l.GoogleToken == "" {
		return errLoginGoogleMalformed
	}

	return nil
}

var errLoginGoogleMalformed = errors.New("token missing")

func LoginGoogleEndpoint(db database.UserAccountInterface, sessionAuth *jwtauth.JWTAuth, gValidator googlehelper.GTokenValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := r.Context()
		data := &loginGoogleRequest{}
		if err := render.Bind(r, data); err != nil {
			log.Debug().Err(err).Msg("Google token body malformed")
			render.Render(w, r, BadRequestError(err))
			return
		}

		gClaims, err := gValidator.ValidateGToken(ctx, data.GoogleToken)
		if err != nil {
			log.Debug().Err(err).Str("token", data.GoogleToken).Msg("Google token validation failed")
			render.Render(w, r, ValidationFailedError(ErrLoginFailed))
			return
		}

		session, err := db.LoginGoogle(ctx, gClaims.Email, gClaims.AccountId)
		if err != nil {
			log.Debug().Err(err).Str("username", gClaims.Email).Msg("Login attempt failed")
			if err == database.ErrAccountNotActive {
				render.Render(w, r, UnauthorizedRequestError(err))
			} else {
				render.Render(w, r, ValidationFailedError(ErrLoginFailed))
			}
			return
		}

		t, tokenString, err := CreateNewSessionToken(sessionAuth, tokenClaimsSchema{gClaims.Email, session})
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
