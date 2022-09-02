package sessiontoken

import (
	"context"
	"ic-rhadi/e_library/database"
	"time"

	"github.com/lestrrat-go/jwx/jwt"
)

func Validator(ctx context.Context, db database.UserSessionInterface, t jwt.Token) error {
	var cs TokenClaimsSchema
	m, err := t.AsMap(ctx)
	if err != nil {
		return err
	}
	if err := cs.StrictFromInterface(m); err != nil {
		return err
	}

	isValid, err := db.CheckSession(ctx, cs.Email, cs.Session, time.Now())
	if err != nil {
		return err
	}
	if !isValid {
		return database.ErrSessionInvalid
	}
	return nil
}
