package endpoints

import (
	"errors"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/emailhelper"
	"net/http"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type registerPostRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var emailRegex = regexp.MustCompile(`^.*@.*\..*$`)
var passwordNumRegex = regexp.MustCompile(`[0-9]`)
var passwordUpperRegex = regexp.MustCompile(`[A-Z]`)
var passwordSpecialRegex = regexp.MustCompile(`[^a-zA-Z0-9]`)

func (l *registerPostRequest) Bind(r *http.Request) error {
	if l.Name == "" || l.Email == "" || l.Password == "" {
		return errRegisterPostMalformed
	}

	if len(l.Email) > 254 || !emailRegex.MatchString(l.Email) {
		return errEmailMalformed
	}

	pLen := utf8.RuneCountInString(l.Password)
	if pLen < 8 {
		return errPasswordTooShort
	}
	if pLen > 40 {
		return errPasswordTooLong
	}

	if !passwordNumRegex.MatchString(l.Password) {
		return errPasswordDontHaveNumber
	}

	if !passwordUpperRegex.MatchString(l.Password) {
		return errPasswordDontHaveUppercase
	}

	if !passwordSpecialRegex.MatchString(l.Password) {
		return errPasswordDontHaveSpecials
	}

	return nil
}

type registerResponse struct {
	NewID string `json:"new_id"`
}

func (reg *registerResponse) Render(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("content-type", "application/json")
	return nil
}

var errRegisterPostMalformed = errors.New("email, password, or name missing")
var errEmailMalformed = errors.New("email form unrecognizable")
var errPasswordTooShort = errors.New("password is too short")
var errPasswordTooLong = errors.New("password is too long")
var errPasswordDontHaveNumber = errors.New("password don't have number")
var errPasswordDontHaveUppercase = errors.New("password don't have uppercase english unaccented latin letters")
var errPasswordDontHaveSpecials = errors.New("password don't have special characters")
var errAccountAlreadyRegistered = errors.New("this account is already registered")

func RegisterPost(
	db database.UserAccountInterface,
	email emailhelper.ActivationMailDriver,
	activationDuration time.Duration,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		data := &registerPostRequest{}
		if err := render.Bind(r, data); err != nil {
			log.Debug().Err(err).Str("email", data.Email).Str("name", data.Name).Msg("Registering attempt failed")
			render.Render(w, r, BadRequestError(err))
			return
		}

		activationToken, validUntil, err := db.Register(ctx, data.Email, data.Password, data.Name, activationDuration)
		if err != nil {
			if err == database.ErrAccountExisted {
				render.Render(w, r, RequestConflictError(errAccountAlreadyRegistered))
				return
			}
			log.Debug().Err(err).Str("email", data.Email).Msg("Registering failed")
			render.Render(w, r, InternalServerError())
			return
		}

		if err := email.SendActivationEmail(data.Email, activationToken, *validUntil); err != nil {
			log.Error().Err(err).
				Str("email", data.Email).
				Str("Activation Token", activationToken).
				Msg("Sending account activation email failed")
		}

		render.Render(w, r, &registerResponse{NewID: data.Email})
	}
}
