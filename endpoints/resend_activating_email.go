package endpoints

import (
	"errors"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/emailhelper"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type resendResponse struct {
	Message string `json:"message"`
}

func (b *resendResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusOK)
	w.Header().Set("content-type", "application/json")
	return nil
}

var errResendActivationEmailMalformed = errors.New("email or token query missing")

func ResendActivationEmail(
	db database.UserAccountInterface,
	email emailhelper.ActivationMailDriver,
	activationDuration time.Duration,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		accountEmail := r.URL.Query().Get("email")
		if accountEmail == "" {
			log.Debug().Msg("Resend activation email endpoint called with insufficient queries")
			render.Render(w, r, BadRequestError(errResendActivationEmailMalformed))
			return
		}

		activationToken, validUntil, err := db.RefreshActivation(ctx, accountEmail, activationDuration)
		if err != nil {
			if err == database.ErrAccountNotFound {
				log.Debug().Err(err).Msg("Resending activating email on an account that couldn't be found")
				render.Render(w, r, UnauthorizedRequestError(errAccountNotFound))
				return
			}
			if err == database.ErrAccountAlreadyActivated {
				log.Debug().Err(err).Msg("Resending activating email on an account that's already activated")
				render.Render(w, r, UnauthorizedRequestError(errAccountAlreadyActivated))
				return
			}
			log.Error().Err(err).Msg("Database error while trying to resend activating email")
			render.Render(w, r, InternalServerError())
			return
		}

		if err := email.SendActivationEmail(accountEmail, activationToken, *validUntil); err != nil {
			log.Debug().Err(err).Msg("Trying to resend activating email")
			render.Render(w, r, InternalServerError())
			return
		}

		resp := resendResponse{"resend succesful"}
		render.Render(w, r, &resp)
	}
}
