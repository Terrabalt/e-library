package endpoints

import (
	"errors"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/googlehelper"
	"ic-rhadi/e_library/sessiontoken"
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
		return ErrLoginGoogleMalformed
	}

	return nil
}

var ErrLoginGoogleMalformed = errors.New("token missing")

func LoginGoogle(
	db database.UserAccountInterface,
	sessionAuth *jwtauth.JWTAuth,
	gValidator googlehelper.GTokenValidator,
	sessionLength time.Duration,
	tokenLength time.Duration,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		data := &loginGoogleRequest{}
		if err := render.Bind(r, data); err != nil {
			log.Debug().Err(err).Msg("Google token body malformed")
			render.Render(w, r, BadRequestError(ErrLoginGoogleMalformed))
			return
		}

		gClaims, err := gValidator.ValidateGToken(ctx, data.GoogleToken)
		if err != nil {
			log.Debug().Err(err).Str("token", data.GoogleToken).Msg("Google token validation failed")
			render.Render(w, r, ValidationFailedError(ErrLoginFailed))
			return
		}

		session, err := db.LoginGoogle(ctx, gClaims.Email, gClaims.AccountID, sessionLength)
		if err != nil {
			switch err {
			case database.ErrAccountNotActive:
				render.Render(w, r, UnauthorizedRequestError(ErrLoginAccountNotActive))
			case database.ErrAccountNotFound, database.ErrWrongID:
				render.Render(w, r, ValidationFailedError(ErrLoginFailed))
			default:
				log.Debug().Err(err).Str("email", gClaims.Email).Msg("Login attempt failed")
				render.Render(w, r, InternalServerError())
			}
			return
		}

		aT, accessTokenString, err := sessiontoken.CreateNewSessionToken(
			sessionAuth,
			sessiontoken.AccessClaimsSchema{
				Email: gClaims.Email,
			},
			tokenLength,
		)
		if err != nil {
			log.Error().Err(err).Msg("Error encoding new session token")
			render.Render(w, r, InternalServerError())
			return
		}

		_, refreshTokenString, err := sessiontoken.CreateNewRefreshToken(
			sessionAuth,
			sessiontoken.RefreshClaimsSchema{
				Email:   gClaims.Email,
				Session: session,
			},
			sessionLength,
		)
		if err != nil {
			log.Error().Err(err).Msg("Error encoding new refresh token")
			render.Render(w, r, InternalServerError())
			return
		}

		render.Render(w, r, &tokenResponse{
			Session:   accessTokenString,
			Refresh:   refreshTokenString,
			Scheme:    "Bearer",
			ExpiresAt: aT.Expiration().Format(time.RFC3339),
		})
	}
}
