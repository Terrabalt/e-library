package endpoints

import (
	"errors"
	"ic-rhadi/e_library/database"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type activatedResponse struct {
	Message string `json:"message"`
	Email   string `json:"email"`
}

func (b *activatedResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusOK)
	w.Header().Set("content-type", "application/json")
	return nil
}

var errAccountNotFound = errors.New("account not found")
var errAccountAlreadyActivated = errors.New("account has already been activated")
var errAccountActivationFailed = errors.New("account activation failed. either the link is invalid or it has expired")
var errAccountActivationQueryMalformed = errors.New("email or token query missing")

func ActivateAccount(
	db database.UserAccountInterface,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		accountEmail := r.URL.Query().Get("email")
		activationToken := r.URL.Query().Get("token")
		if accountEmail == "" || activationToken == "" {
			log.Debug().Msg("activating account endpoint called with insufficient queries")
			render.Render(w, r, BadRequestError(errAccountActivationQueryMalformed))
			return
		}

		activated, token, expiresIn, err := db.GetActivationData(ctx, accountEmail)
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
		if activationToken == "" || expiresIn == nil {
			log.Error().Str("account", accountEmail).Msg("trying to activate account returned an unexpected error")
			render.Render(w, r, InternalServerError())
			return
		}
		if activationToken != token || expiresIn.Before(time.Now()) {
			log.Debug().Err(err).Msg("trying to activate account failed")
			render.Render(w, r, UnauthorizedRequestError(errAccountActivationFailed))
			return
		}

		if err := db.ActivateAccount(ctx, accountEmail); err != nil {
			log.Error().Err(err).Msg("trying to activate account returned an unexpected error")
			render.Render(w, r, InternalServerError())
			return
		}

		resp := activatedResponse{
			Message: "account activated",
			Email:   accountEmail,
		}
		render.Render(w, r, &resp)
	}
}
