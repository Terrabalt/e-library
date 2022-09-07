package endpoints

import (
	"errors"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/emailhelper"
	"ic-rhadi/e_library/googlehelper"
	"net/http"
	"time"

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

var errRegisterGoogleMalformed = errors.New("token missing")
var errGoogleTokenFailed = errors.New("google token validation failed")

func RegisterGoogle(
	db database.UserAccountInterface,
	gValidator googlehelper.GTokenValidator,
	email emailhelper.ActivationMailDriver,
	activationDuration time.Duration,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := r.Context()

		data := &registerGoogleRequest{}
		if err := render.Bind(r, data); err != nil {
			log.Debug().Err(err).Msg("Registering attempt failed")
			render.Render(w, r, BadRequestError(err))
			return
		}

		gClaims, err := gValidator.ValidateGToken(ctx, data.GoogleToken)
		if err != nil {
			log.Debug().Err(err).Str("token", data.GoogleToken).Msg("Google token validation failed")
			render.Render(w, r, ValidationFailedError(errGoogleTokenFailed))
			return
		}

		activationToken, validUntil, err := db.RegisterGoogle(ctx, gClaims.Email, gClaims.AccountID, gClaims.FullName, activationDuration)
		if err != nil {
			if err == database.ErrAccountExisted {
				render.Render(w, r, RequestConflictError(errAccountAlreadyRegistered))
				return
			}
			log.Debug().Err(err).Str("email", gClaims.Email).Msg("Registering failed")
			render.Render(w, r, InternalServerError())
			return
		}

		if err := email.SendActivationEmail(gClaims.Email, activationToken, *validUntil); err != nil {
			log.Error().Err(err).
				Str("email", gClaims.Email).
				Str("Activation Token", activationToken).
				Msg("Sending account activation email failed")
		}

		render.Render(w, r, &registerResponse{NewID: gClaims.Email})
	}
}
