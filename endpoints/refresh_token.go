package endpoints

import (
	"errors"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/sessiontoken"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/google/uuid"
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

var errRefreshTokenMalformed = errors.New("token missing from request body")
var errRefreshTokenInvalid = errors.New("invalid refresh token")
var errRefreshTokenExpired = errors.New("refresh token expired")

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
			log.Debug().Err(err).Msg("Refresh token endpoint called with insufficient body")
			render.Render(w, r, BadRequestError(errRefreshTokenMalformed))
			return
		}

		currToken, err := sessionAuth.Decode(data.RefreshToken)
		if err != nil {
			log.Debug().Err(err).Msg("Trying to decode refresh token returned an error")
			render.Render(w, r, UnauthorizedRequestError(errRefreshTokenInvalid))
			return
		}

		var claims sessiontoken.RefreshClaimsSchema
		if err := claims.FromToken(ctx, currToken); err != nil {
			log.Debug().Err(err).Msg("Trying to get refresh token's claims returned an error")
			render.Render(w, r, UnauthorizedRequestError(errRefreshTokenInvalid))
			return
		}

		currTime := time.Now()
		tokenFamily, exhausted, expiresIn, err := db.GetSession(ctx, claims.Email, claims.Session, currTime)
		if err != nil {
			if err == database.ErrSessionNotFound {
				log.Debug().Msg("")
				render.Render(w, r, UnauthorizedRequestError(errRefreshTokenInvalid))
				return
			}
			log.Error().Err(err).Msg("Database error while trying to get ")
			render.Render(w, r, InternalServerError())
			return
		}

		if expiresIn.Before(currTime) {
			log.Debug().Msg("Refresh token has expired")
			render.Render(w, r, UnauthorizedRequestError(errRefreshTokenExpired))
			return
		}
		if exhausted {
			if err := db.InvaildateSession(ctx, claims.Email, claims.Session); err != nil {
				log.Error().Err(err).Msg("Database error while invalidating session tokens")
				render.Render(w, r, InternalServerError())
				return
			}
			log.Debug().Str("account", claims.Email).Msg("Double refresh of the same token. Token family invalidated")
			render.Render(w, r, UnauthorizedRequestError(errRefreshTokenExpired))
			return
		}

		newSession, err := uuid.NewRandom()
		if err != nil {
			log.Error().Err(err).Msg("Unexpected error while creating new random UUID for token")
			render.Render(w, r, InternalServerError())
			return
		}
		if err := db.AddNewSession(ctx, claims.Email, newSession.String(), tokenFamily, currTime.Add(sessionLength)); err != nil {
			log.Error().Err(err).Msg("Database error while adding a new refresh token")
			render.Render(w, r, InternalServerError())
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
				Session: newSession.String(),
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
