package endpoints

import (
	"context"
	"errors"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/sessiontoken"
	"net/http"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/rs/zerolog/log"
)

var ErrSessionTokenMissingOrInvalid = errors.New("session token invalid or missing")

func SessionAuthenticatorMiddleware(db database.UserSessionInterface) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// Token has already been verified as current and signed properly from previous
			// middleware jwtauth.Verifier(ja *jwtauth.JWTAuth)
			token, _, err := jwtauth.FromContext(ctx)

			if err != nil {
				log.Debug().Err(err).Msg("Getting the jwt token returned an error")
				render.Render(w, r, UnauthorizedRequestError(ErrSessionTokenMissingOrInvalid))
				return
			}

			if token == nil {
				log.Debug().Msg("Getting the jwt token returned null")
				render.Render(w, r, UnauthorizedRequestError(ErrSessionTokenMissingOrInvalid))
				return
			}

			ctx = context.WithValue(ctx, sessiontoken.DatabaseContextKey{}, db)
			if err := jwt.Validate(
				token,
				jwt.WithContext(ctx),
				jwt.WithValidator(jwt.ValidatorFunc(sessiontoken.Validator)),
			); err != nil {
				if err == database.ErrSessionInvalid {
					log.Debug().Msg("Validating jwt token returned failure")
					render.Render(w, r, UnauthorizedRequestError(ErrSessionTokenMissingOrInvalid))
				} else {
					log.Error().Err(err).Msg("Validating session token returned an error")
					render.Render(w, r, InternalServerError())
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
