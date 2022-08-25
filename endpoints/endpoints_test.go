package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"ic-rhadi/e_library/googlehelper"
	"ic-rhadi/e_library/sessiontoken"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/stretchr/testify/mock"
)

var expID = struct {
	Account  string
	Session  string
	GAccount string
}{
	"123456778-9abc-def0@1234.56789abcdef0",
	"12345678-9abc-def0-1234-56789abcdef0",
	"123456789abcdef0123456789abcdef0",
}

type param struct {
	ID    string
	Param string
}

var tokenAuth = jwtauth.New("HS256", []byte("secret"), nil)

const expSessionLen = time.Hour * time.Duration(48)
const expTokenLen = time.Minute * time.Duration(10)

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
		routerCtx.URLParams.Add(p.ID, p.Param)
	}
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, routerCtx)

	if withToken {
		expToken, _, err := constToken(t, tokenAuth, expID.Account, expID.Session)
		ctx = jwtauth.NewContext(ctx, expToken, err)
	}

	r = r.WithContext(ctx)
	return r, w
}

func constToken(t *testing.T, tokenAuth *jwtauth.JWTAuth, email, session string) (jwt.Token, string, error) {
	expClaims, err := sessiontoken.TokenClaimsSchema{
		Email:   email,
		Session: session,
	}.ToInterface()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when making mock data", err)
		return nil, "", err
	}
	expToken, expTokenString, err := tokenAuth.Encode(expClaims)
	return expToken, expTokenString, err
}

func (e ErrorResponse) sentForm() (ErrorResponse, int) {
	status := e.httpStatusCode
	e.httpStatusCode = 0
	return e, status
}

type dBMock struct {
	mock.Mock
}

func (db dBMock) Login(ctx context.Context, email string, pass string, sessionLength time.Duration) (id string, err error) {
	args := db.Called(email, pass, sessionLength)
	return args.String(0), args.Error(1)
}

func (db dBMock) LoginGoogle(ctx context.Context, gID string, pass string, sessionLength time.Duration) (id string, err error) {
	args := db.Called(gID, pass, sessionLength)
	return args.String(0), args.Error(1)
}

func (db dBMock) Register(ctx context.Context, email string, password string, name string, viaGoogle bool) error {
	args := db.Called(email, password, name, viaGoogle)
	return args.Error(0)
}

type gTokenValidatorMock struct {
	mock.Mock
}

func (g gTokenValidatorMock) ValidateGToken(ctx context.Context, token string) (*googlehelper.GoogleClaimsSchema, error) {
	args := g.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*googlehelper.GoogleClaimsSchema), args.Error(1)
}
