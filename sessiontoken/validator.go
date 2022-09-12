package sessiontoken

import (
	"context"

	"github.com/lestrrat-go/jwx/jwt"
)

func Validator(ctx context.Context, t jwt.Token) error {
	var cs TokenClaimsSchema
	m, err := t.AsMap(ctx)
	if err != nil {
		return err
	}
	if err := cs.StrictFromInterface(m); err != nil {
		return err
	}
	return nil
}
