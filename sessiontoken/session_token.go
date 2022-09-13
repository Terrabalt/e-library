package sessiontoken

import (
	"context"
	"encoding/json"
	"errors"
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

func (token *TokenClaimsSchema) StrictFromInterface(inter map[string]interface{}) error {
	if err := token.FromInterface(inter); err != nil {
		return err
	}
	return token.CheckMalform()
}

func (token *TokenClaimsSchema) StrictToInterface() (inter map[string]interface{}, err error) {
	if err := token.CheckMalform(); err != nil {
		return nil, err
	}
	return token.ToInterface()
}

var ErrTokenMalformed = errors.New("important data missing from token schema")

func (token TokenClaimsSchema) CheckMalform() error {
	if token.Email == "" || token.Session == "" {
		return ErrTokenMalformed
	}
	return nil
}

func CreateNewSessionToken(tokenAuth *jwtauth.JWTAuth, claims TokenClaimsSchema, tokenLength time.Duration) (token jwt.Token, tokenString string, err error) {
	c, err := claims.ToInterface()
	if err != nil {
		return nil, "", err
	}

	now := time.Now()
	jwtauth.SetIssuedAt(c, now)
	c[jwt.NotBeforeKey] = now.UTC().Unix()
	jwtauth.SetExpiryIn(c, tokenLength)

	return tokenAuth.Encode(c)
}

func FromContext(ctx context.Context) (*TokenClaimsSchema, error) {
	_, token, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	var sch TokenClaimsSchema
	if err := sch.FromInterface(token); err != nil {
		return nil, err
	}
	return &sch, nil
}
