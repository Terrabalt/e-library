package googlehelper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
)

type GTokenValidator interface {
	ValidateGToken(ctx context.Context, token string) (*GoogleClaimsSchema, error)
}
type GTokenValidatorInst struct {
	tokenValidator *idtoken.Validator
}

func NewGValidator(ctx context.Context) (GTokenValidator, error) {
	idTokenVal, err := idtoken.NewValidator(ctx, option.WithoutAuthentication())
	if err != nil {
		return nil, fmt.Errorf("validator creation failed: %w", err)
	}
	return &GTokenValidatorInst{
		tokenValidator: idTokenVal,
	}, nil
}

type GoogleClaimsSchema struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FullName      string `json:"name"`
	AccountId     string `json:"sub"`
}

var ErrGoogleEmailUnverified = errors.New("google email unverified")

func (v *GTokenValidatorInst) ValidateGToken(ctx context.Context, token string) (*GoogleClaimsSchema, error) {
	tkn, err := v.tokenValidator.Validate(ctx, token, "")
	if err != nil {
		return nil, err
	}

	tknMap := tkn.Claims
	claims := &GoogleClaimsSchema{}
	js, err := json.Marshal(tknMap)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(js, claims)
	if err != nil {
		return nil, err
	}
	if !claims.EmailVerified {
		return nil, ErrGoogleEmailUnverified
	}

	return claims, nil
}
