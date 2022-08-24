package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/stretchr/testify/mock"
)

var expId = struct {
	Account  string
	Session  string
	GAccount string
}{
	"123456778-9abc-def0@1234.56789abcdef0",
	"12345678-9abc-def0-1234-56789abcdef0",
	"123456789abcdef0123456789abcdef0",
}

type param struct {
	Id    string
	Param string
}

var tokenAuth = jwtauth.New("HS256", []byte("secret"), nil)

func mockRequest(t *testing.T, path string, body interface{}, withToken bool, params ...param) (*http.Request, *httptest.ResponseRecorder) {
	var r *http.Request

	var req io.Reader = nil
	if body != nil {
		js, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("an error '%s' was not expected when making mock data", err)
		}

		req = bytes.NewBuffer(js)
	}
	r = httptest.NewRequest(http.MethodPost, path, req)
	w := httptest.NewRecorder()
	r.Header.Add("Content-Type", "application/json")

	routerCtx := chi.NewRouteContext()
	for _, p := range params {
		routerCtx.URLParams.Add(p.Id, p.Param)
	}
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, routerCtx)

	if withToken {
		expToken, _, err := constToken(t, tokenAuth, expId.Account, expId.Session)
		ctx = jwtauth.NewContext(ctx, expToken, err)
	}

	r = r.WithContext(ctx)
	return r, w
}

func constToken(t *testing.T, tokenAuth *jwtauth.JWTAuth, email, session string) (jwt.Token, string, error) {
	expClaims, err := tokenClaimsSchema{
		Email:   email,
		Session: session,
	}.toInterface()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when making mock data", err)
		return nil, "", err
	}
	expToken, expTokenString, err := tokenAuth.Encode(expClaims)
	return expToken, expTokenString, err
}

func (e ErrorResponse) sentForm() (ErrorResponse, int) {
	e.err = nil
	status := e.httpStatusCode
	e.httpStatusCode = 0
	return e, status
}

type DBMock struct {
	mock.Mock
}

func (db DBMock) Login(ctx context.Context, email string, pass string, viaGoogle bool) (id string, err error) {
	args := db.Called(email, pass, viaGoogle)
	return args.String(0), args.Error(1)
}

type gTokenValidatorMock struct {
	mock.Mock
}

func (g gTokenValidatorMock) validateGToken(ctx context.Context, token string) (*googleClaimsSchema, error) {
	args := g.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*googleClaimsSchema), args.Error(1)
}
