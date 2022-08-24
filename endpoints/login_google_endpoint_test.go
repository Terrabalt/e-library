package endpoints

import (
	"encoding/json"
	"errors"
	"ic-rhadi/e_library/database"
	"net/http"
	"testing"

	"github.com/go-chi/jwtauth"
	"github.com/stretchr/testify/assert"
)

func TestSuccessfulLoginGoogle(t *testing.T) {
	path := "/auth/login"

	login := loginGoogleRequest{
		GoogleToken: "a.b.c",
	}

	expGClaims := googleClaimsSchema{
		Email:         expId.Account,
		EmailVerified: true,
		FullName:      "Joko",
		AccountId:     expId.GAccount,
	}

	gValidatorMock := &gTokenValidatorMock{}
	gValidatorMock.On("validateGToken", login.GoogleToken).
		Return(&expGClaims, nil).Once()

	dbMock := &DBMock{}
	dbMock.On("Login", expGClaims.Email, expGClaims.AccountId, true).
		Return(expId.Session, nil).Once()

	expCode := http.StatusOK

	r, w := mockRequest(t, path, login, false)
	handler := LoginGoogleEndpoint(dbMock, tokenAuth, gValidatorMock)
	handler.ServeHTTP(w, r)

	resp := &tokenResponse{}
	assert.Equal(t, expCode, w.Code, "A successful Google-Login didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful Google-Login didn't return a valid tokenResponse object")
	token, err := jwtauth.VerifyToken(tokenAuth, resp.Token)
	assert.NoError(t, err, "A successful Google-Login didn't return a valid token")

	email, _ := token.Get("email")
	session, _ := token.Get("session")
	assert.Equal(t, expId.Account, email, "A successful Google-Login didn't return expected email")
	assert.Equal(t, expId.Session, session, "A successful Google-Login didn't return expected session id")
	assert.True(t, gValidatorMock.AssertExpectations(t), "The test has unfulfilled expectation")
	assert.True(t, dbMock.AssertExpectations(t), "The test has unfulfilled expectation")
}

func TestMalformedLoginGoogle(t *testing.T) {
	path := "/auth/login"

	login := loginPostRequest{
		Username: "",
		Password: "Password",
	}

	gValidatorMock := &gTokenValidatorMock{}
	dbMock := &DBMock{}

	expResp, expCode := BadRequestError(errLoginGoogleMalformed).(*ErrorResponse).
		sentForm()

	r, w := mockRequest(t, path, login, false)
	handler := LoginGoogleEndpoint(dbMock, tokenAuth, gValidatorMock)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A malformed Google-Login didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A malformed Google-Login didn't return a valid errorResponse object")
	assert.Equal(t, expResp, *resp, "A malformed Google-Login didn't return the proper error")
	assert.True(t, gValidatorMock.AssertExpectations(t), "The test has unfulfilled expectation")
	assert.True(t, dbMock.AssertExpectations(t), "The test has unfulfilled expectation")
}
func TestFailedLoginGoogle(t *testing.T) {

	path := "/auth/login"

	login := loginGoogleRequest{
		GoogleToken: "a.b.c",
	}

	expGClaims := googleClaimsSchema{
		Email:         expId.Account,
		EmailVerified: true,
		FullName:      "Joko",
		AccountId:     expId.GAccount,
	}

	gValidatorMock := &gTokenValidatorMock{}
	gValidatorMock.On("validateGToken", login.GoogleToken).
		Return(nil, errors.New("password wrong, should be xxxxx")).Once()

	dbMock := &DBMock{}

	expResp, expCode := ValidationFailedError(ErrLoginFailed).(*ErrorResponse).
		sentForm()

	r, w := mockRequest(t, path, login, false)
	handler := LoginGoogleEndpoint(dbMock, tokenAuth, gValidatorMock)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A failed Google-Login didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A failed Google-Login didn't return a valid errorResponse object")
	assert.Equal(t, expResp, *resp, "A failed Google-Login didn't return the proper error")
	assert.True(t, gValidatorMock.AssertExpectations(t), "The test has unfulfilled expectation")
	assert.True(t, dbMock.AssertExpectations(t), "The test has unfulfilled expectation")

	gValidatorMock.On("validateGToken", login.GoogleToken).
		Return(&expGClaims, nil).Once()

	dbMock.On("Login", expGClaims.Email, expGClaims.AccountId, true).
		Return("", database.ErrAccountNotFound).Once()

	r, w = mockRequest(t, path, login, false)
	handler = LoginGoogleEndpoint(dbMock, tokenAuth, gValidatorMock)
	handler.ServeHTTP(w, r)

	resp = &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A failed Google-Login didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A failed Google-Login didn't return a valid errorResponse object")
	assert.Equal(t, expResp, *resp, "A failed Google-Login didn't return the proper error")
	assert.True(t, gValidatorMock.AssertExpectations(t), "The test has unfulfilled expectation")
	assert.True(t, dbMock.AssertExpectations(t), "The test has unfulfilled expectation")
}

func TestNotActivatedLoginGoogle(t *testing.T) {
	path := "/auth/login"

	login := loginGoogleRequest{
		GoogleToken: "a.b.c",
	}

	expGClaims := googleClaimsSchema{
		Email:         expId.Account,
		EmailVerified: true,
		FullName:      "Joko",
		AccountId:     expId.GAccount,
	}

	gValidatorMock := &gTokenValidatorMock{}
	gValidatorMock.On("validateGToken", login.GoogleToken).
		Return(&expGClaims, nil).Once()

	dbMock := &DBMock{}
	dbMock.On("Login", expGClaims.Email, expGClaims.AccountId, true).
		Return("", database.ErrAccountNotActive).Once()

	expResp, expCode := UnauthorizedRequestError(database.ErrAccountNotActive).(*ErrorResponse).
		sentForm()

	r, w := mockRequest(t, path, login, false)
	handler := LoginGoogleEndpoint(dbMock, tokenAuth, gValidatorMock)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A Google-Login on a not activated account didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A Google-Login on a not activated account didn't return a valid errorResponse object")
	assert.Equal(t, expResp, *resp, "A Google-Login on a not activated account didn't return the proper error")
	assert.True(t, gValidatorMock.AssertExpectations(t), "The test has unfulfilled expectation")
	assert.True(t, dbMock.AssertExpectations(t), "The test has unfulfilled expectation")
}
