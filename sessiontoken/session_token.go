package sessiontoken

import (
	"encoding/json"
	"time"

	"github.com/go-chi/jwtauth"
	"github.com/lestrrat-go/jwx/jwt"
)

type TokenClaimsSchema struct {
	Email   string `json:"sub"`
	Session string `json:"session"`
}

func (token TokenClaimsSchema) ToInterface() (inter map[string]interface{}, err error) {
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

func (token *TokenClaimsSchema) FromInterface(inter map[string]interface{}) error {
	js, err := json.Marshal(inter)
	if err != nil {
		return err
	}

	err = json.Unmarshal(js, token)
	if err != nil {
		return err
	}
	return nil
}

func CreateNewSessionToken(tokenAuth *jwtauth.JWTAuth, claims TokenClaimsSchema) (token jwt.Token, tokenString string, err error) {
	c, err := claims.ToInterface()
	if err != nil {
		return nil, "", err
	}

	now := time.Now()
	jwtauth.SetIssuedAt(c, now)
	c["nbf"] = now.UTC().Unix()
	jwtauth.SetExpiryIn(c, time.Duration(2)*time.Hour)

	return tokenAuth.Encode(c)
}
