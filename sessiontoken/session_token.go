package sessiontoken

import (
	"context"
	"errors"
	"time"

	"github.com/go-chi/jwtauth"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwt"
)

func CreateNewSessionToken(tokenAuth *jwtauth.JWTAuth, claims AccessClaimsSchema, tokenLength time.Duration) (token jwt.Token, tokenString string, err error) {
	c, err := claims.ToInterface()
	if err != nil {
		return nil, "", err
	}

	now := time.Now()
	jwtauth.SetIssuedAt(c, now)
	c[jwt.NotBeforeKey] = now.UTC().Unix()
	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, "", err
	}
	c[jwt.JwtIDKey] = randomUUID
	jwtauth.SetExpiryIn(c, tokenLength)

	return tokenAuth.Encode(c)
}

func CreateNewRefreshToken(tokenAuth *jwtauth.JWTAuth, claims RefreshClaimsSchema, tokenLength time.Duration) (token jwt.Token, tokenString string, err error) {
	c, err := claims.ToInterface()
	if err != nil {
		return nil, "", err
	}

	now := time.Now()
	jwtauth.SetIssuedAt(c, now)
	c[jwt.NotBeforeKey] = now.UTC().Unix()
	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, "", err
	}
	c[jwt.JwtIDKey] = randomUUID
	jwtauth.SetExpiryIn(c, tokenLength)

	return tokenAuth.Encode(c)
}

var ErrSessionTokenMissing = errors.New("session token missing")

func FromContext(ctx context.Context) (*AccessClaimsSchema, error) {
	_, token, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if len(token) == 0 {
		return nil, ErrSessionTokenMissing
	}
	var sch AccessClaimsSchema
	if err := sch.FromInterface(token); err != nil {
		return nil, err
	}
	return &sch, nil
}
