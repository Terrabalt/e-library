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

type refreshTokenRequest struct {
	RefreshToken string `json:"token"`
}

func (l *refreshTokenRequest) Bind(r *http.Request) error {
	if l.RefreshToken == "" {
		return errRegisterGoogleMalformed
	}

	return nil
}

var ErrRefreshTokenMalformed = errors.New("")

func RefreshToken(
	db database.UserSessionInterface,
	sessionAuth *jwtauth.JWTAuth,
	sessionLength time.Duration,
	tokenLength time.Duration,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		data := &refreshTokenRequest{}
		if err := render.Bind(r, data); err != nil {
			log.Debug().Err(err).Msg("")
			render.Render(w, r, BadRequestError(errors.New("")))
			return
		}

		currToken, err := sessionAuth.Decode(data.RefreshToken)
		if err != nil {
			log.Debug().Err(err).Msg("")
			render.Render(w, r, UnauthorizedRequestError(errors.New("")))
			return
		}

		var claims sessiontoken.RefreshClaimsSchema
		if err := claims.FromToken(ctx, currToken); err != nil {
			log.Debug().Err(err).Msg("")
		}

		newSession, err := db.CheckSession(ctx, claims.Email, claims.Session, time.Now(), sessionLength)
		if err != nil {
			log.Debug().Err(err).Msg("")
			render.Render(w, r, UnauthorizedRequestError(errors.New("")))
			return
		}

		aT, accessTokenString, err := sessiontoken.CreateNewSessionToken(
			sessionAuth,
			sessiontoken.AccessClaimsSchema{
				Email: claims.Email,
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
				Email:   claims.Email,
				Session: newSession,
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
