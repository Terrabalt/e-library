package endpoints

import (
	"errors"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/emailhelper"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
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

		activated, _, _, err := db.GetActivationData(ctx, accountEmail)
		if err == database.ErrAccountNotFound {
			log.Debug().Err(err).Msg("Resending activating email on an account that couldn't be found")
			render.Render(w, r, UnauthorizedRequestError(errAccountNotFound))
			return
		}
		if activated {
			log.Debug().Err(err).Msg("Resending activating email on an account that's already activated")
			render.Render(w, r, UnauthorizedRequestError(errAccountAlreadyActivated))
			return
		}

		activationToken, err := uuid.NewRandom()
		if err != nil {
			log.Error().Err(err).Msg("Trying to create random uuid returned an error")
			render.Render(w, r, InternalServerError())
			return
		}
		expiresIn := time.Now().Add(activationDuration)

		if err := db.RefreshActivation(ctx, accountEmail, activationToken.String(), expiresIn); err != nil {
			log.Error().Err(err).Msg("Database error while trying to refresh activation token")
			render.Render(w, r, InternalServerError())
			return
		}
		if err := email.SendActivationEmail(accountEmail, activationToken.String(), expiresIn); err != nil {
			log.Debug().Err(err).Msg("Trying to resend activating email failed")
			render.Render(w, r, InternalServerError())
			return
		}

		resp := resendResponse{"resend succesful"}
		render.Render(w, r, &resp)
	}
}
