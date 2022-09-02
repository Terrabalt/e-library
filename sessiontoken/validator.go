package sessiontoken

import (
	"context"
	"ic-rhadi/e_library/database"
	"time"

	"github.com/lestrrat-go/jwx/jwt"
)

type DatabaseContextKey struct{}

func Validator(ctx context.Context, t jwt.Token) error {
	var cs TokenClaimsSchema
	m, err := t.AsMap(ctx)
	if err != nil {
		return err
	}
	cs.FromInterface(m)

	db := ctx.Value(DatabaseContextKey{}).(database.UserSessionInterface)
	isValid, err := db.CheckSession(ctx, cs.Email, cs.Session, time.Now())
	if err != nil {
		return err
	}
	if !isValid {
		return database.ErrSessionInvalid
	}
	return nil
}
