package endpoints

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type tokenClaimsSchema struct {
	Email   string `json:"sub"`
	Session string `json:"session"`
}

func (token tokenClaimsSchema) toInterface() (inter map[string]interface{}, err error) {
	js, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(js, &inter)
	if err != nil {
		return nil, err
	}
	return
}

type tokenResponse struct {
	Token     string `json:"token"`
	Scheme    string `json:"scheme"`
	ExpiresAt string `json:"expires_at"`
}

func (token tokenResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusOK)
	w.Header().Set("content-type", "application/json")
	return nil
}

var errTokenCreationFailed = errors.New("token creation failed")

func sendNewToken(tokenAuth *jwtauth.JWTAuth, claims tokenClaimsSchema, w http.ResponseWriter, r *http.Request) {
	c, err := claims.toInterface()
	if err != nil {
		log.Error().Err(err).Caller().Msg("Error making new token")
		render.Render(w, r, InternalServerError(errTokenCreationFailed))
		return
	}

	now := time.Now()
	jwtauth.SetIssuedAt(c, now)
	c["nbf"] = now.UTC().Unix()
	jwtauth.SetExpiryIn(c, time.Duration(2)*time.Hour)

	t, tokenString, err := tokenAuth.Encode(c)
	if err != nil {
		log.Error().Err(err).Caller().Msg("Error encoding new token")
		render.Render(w, r, InternalServerError(errTokenCreationFailed))
		return
	}

	token := tokenResponse{
		Token:     tokenString,
		Scheme:    "Bearer",
		ExpiresAt: t.Expiration().Format(time.RFC3339),
	}
	render.Render(w, r, &token)
}
