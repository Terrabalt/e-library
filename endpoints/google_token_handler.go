package endpoints

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
)

type gTokenValidator interface {
	validateGToken(ctx context.Context, token string) (*googleClaimsSchema, error)
}
type gTokenValidatorInst struct {
	tokenValidator *idtoken.Validator
}

func NewGValidator(ctx context.Context) (gTokenValidator, error) {
	idTokenVal, err := idtoken.NewValidator(ctx, option.WithoutAuthentication())
	if err != nil {
		return nil, fmt.Errorf("validator creation failed: %w", err)
	}
	return &gTokenValidatorInst{
		tokenValidator: idTokenVal,
	}, nil
}

type googleClaimsSchema struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FullName      string `json:"name"`
	AccountId     string `json:"sub"`
}

var errGoogleEmailUnverified = errors.New("google email unverified")

func (v *gTokenValidatorInst) validateGToken(ctx context.Context, token string) (*googleClaimsSchema, error) {
	tkn, err := v.tokenValidator.Validate(ctx, token, "")
	if err != nil {
		return nil, err
	}

	tknMap := tkn.Claims
	claims := &googleClaimsSchema{}
	js, err := json.Marshal(tknMap)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(js, claims)
	if err != nil {
		return nil, err
	}
	if !claims.EmailVerified {
		return nil, errGoogleEmailUnverified
	}

	return claims, errors.New("endpoint not implemented")
}
