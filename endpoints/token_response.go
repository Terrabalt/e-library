package endpoints

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/lestrrat-go/jwx/jwt"
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

func CreateNewSessionToken(tokenAuth *jwtauth.JWTAuth, claims tokenClaimsSchema) (token jwt.Token, tokenString string, err error) {
	c, err := claims.toInterface()
	if err != nil {
		return nil, "", err
	}

	now := time.Now()
	jwtauth.SetIssuedAt(c, now)
	c["nbf"] = now.UTC().Unix()
	jwtauth.SetExpiryIn(c, time.Duration(2)*time.Hour)

	return tokenAuth.Encode(c)
}
