package sessiontoken

import (
	"context"

	"github.com/lestrrat-go/jwx/jwt"
)

func Validator(ctx context.Context, t jwt.Token) error {
	m, err := t.AsMap(ctx)
	if err != nil {
		return err
	}

	var cs AccessClaimsSchema
	if err := cs.FromInterface(m); err != nil {
		return err
	}
	return cs.CheckMalform()
}
