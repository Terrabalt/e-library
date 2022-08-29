package endpoints

import (
	"errors"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/emailhelper"
	"ic-rhadi/e_library/googlehelper"
	"net/http"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type registerGoogleRequest struct {
	GoogleToken string `json:"token"`
}

func (l *registerGoogleRequest) Bind(r *http.Request) error {
	if l.GoogleToken == "" {
		return errRegisterGoogleMalformed
	}

	return nil
}

var errRegisterGoogleMalformed = errors.New("")

func RegisterGoogle(
	db database.UserAccountInterface,
	sessionAuth *jwtauth.JWTAuth,
	gValidator googlehelper.GTokenValidator,
	email emailhelper.ActivationMailDriver,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := r.Context()

		defer func() {
			if rec := recover(); rec != nil {
				log.Debug().Str("recovered", rec.(string)).Msg("Panicked while trying to register")
				render.Render(w, r, InternalServerError())
			}
		}()

		data := &registerGoogleRequest{}
		if err := render.Bind(r, data); err != nil {
			log.Debug().Err(err).Msg("Registering attempt failed")
			render.Render(w, r, BadRequestError(err))
			return
		}

		gClaims, err := gValidator.ValidateGToken(ctx, data.GoogleToken)
		if err != nil {
			log.Debug().Err(err).Str("token", data.GoogleToken).Msg("Google token validation failed")
			render.Render(w, r, ValidationFailedError(ErrLoginFailed))
			return
		}

		activationToken, validUntil, err := db.RegisterGoogle(ctx, gClaims.Email, gClaims.AccountId, gClaims.FullName)
		if err != nil {
			log.Debug().Err(err).Str("email", gClaims.Email).Msg("Registering failed")
			render.Render(w, r, InternalServerError())
			return
		}

		if err := email.SendActivationEmail(w, r, gClaims.Email, activationToken, *validUntil); err != nil {
			log.Debug().Err(err).Str("email", gClaims.Email).Str("Activation Token", activationToken).Msg("Registering failed")
			render.Render(w, r, InternalServerError())
			return
		}

		r.Response.StatusCode = http.StatusCreated
	}
}
