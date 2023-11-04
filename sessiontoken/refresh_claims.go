package sessiontoken

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/lestrrat-go/jwx/jwt"
)

type RefreshClaimsSchema struct {
	Email   string `json:"sub"`
	Session string `json:"session"`
}

func (token RefreshClaimsSchema) ToInterface() (inter map[string]interface{}, err error) {
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

func (token *RefreshClaimsSchema) FromInterface(inter map[string]interface{}) error {
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

func (token *RefreshClaimsSchema) FromToken(ctx context.Context, jt jwt.Token) error {
	tokenMap, err := jt.AsMap(ctx)
	if err != nil {
		return err
	}
	return token.FromInterface(tokenMap)
}

var ErrRefreshTokenMalformed = errors.New("important data missing from refresh token schema")

func (token RefreshClaimsSchema) CheckMalform() error {
	if token.Email == "" || token.Session == "" {
		return ErrRefreshTokenMalformed
	}
	return nil
}
