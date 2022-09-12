package endpoints

import (
	"errors"
	"ic-rhadi/e_library/sessiontoken"
	"net/http"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

var ErrSessionTokenMissingOrInvalid = errors.New("session token has expired or missing")

func SessionAuthenticatorMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// Token has already been verified (successfully or failed) as current and signed
			// properly from previous middleware jwtauth.Verifier(ja *jwtauth.JWTAuth).
			// Output has not been yet.
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

			if err := sessiontoken.Validator(ctx, token); err != nil {
				log.Debug().Err(err).Msg("Validating jwt token returned failure")
				render.Render(w, r, UnauthorizedRequestError(ErrSessionTokenMissingOrInvalid))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
