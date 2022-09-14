package endpoints

import (
	"context"
	"encoding/json"
	"errors"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/googlehelper"
	"ic-rhadi/e_library/sessiontoken"
	"net/http"
	"testing"

	"github.com/go-chi/jwtauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSuccessfulLoginGoogle(t *testing.T) {
	path := "/auth/google"

	login := loginGoogleRequest{
		GoogleToken: "a.b.c",
	}

	expGClaims := googlehelper.GoogleClaimsSchema{
		Email:         expID.Account,
		EmailVerified: true,
		FullName:      "Joko",
		AccountID:     expID.GAccount,
	}

	gValidatorMock := gTokenValidatorMock{&mock.Mock{}}
	gValidatorMock.On("ValidateGToken", login.GoogleToken).
		Return(&expGClaims, nil).Once()

	dbMock := dBMock{&mock.Mock{}}
	dbMock.On("LoginGoogle", expGClaims.Email, expGClaims.AccountID, expSessionLen).
		Return(expID.Session, nil).Once()

	expCode := http.StatusOK
	expClaims := sessiontoken.AccessClaimsSchema{
		Email: expID.Account,
	}

	w, r := mockRequest(t, path, login, false)
	handler := LoginGoogle(dbMock, tokenAuth, gValidatorMock, expSessionLen, expTokenLen)
	handler.ServeHTTP(w, r)

	resp := &tokenResponse{}
	assert.Equal(t, expCode, w.Code, "A successful Google-Login didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful Google-Login didn't return a valid tokenResponse object") {
		token, err := jwtauth.VerifyToken(tokenAuth, resp.Token)
		assert.NoError(t, err, "A successful Google-Login didn't return a valid token")

		tokenMap, err := token.AsMap(context.Background())
		if assert.NoErrorf(t, err, "an error '%s' was not expected when getting returned token's schema") {
			var claims sessiontoken.AccessClaimsSchema
			err := claims.FromInterface(tokenMap)
			if assert.NoErrorf(t, err, "an error '%s' was not expected when getting returned token's schema") {
				assert.Equal(t, expClaims, claims, "A successful Post-Login didn't return expected token schema")
			}
		}
	}
	gValidatorMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
}

func TestMalformedLoginGoogle(t *testing.T) {
	path := "/auth/google"

	login := loginPostRequest{
		Email:    "",
		Password: "Password",
	}

	gValidatorMock := gTokenValidatorMock{&mock.Mock{}}
	dbMock := dBMock{&mock.Mock{}}

	expResp, expCode := BadRequestError(ErrLoginGoogleMalformed).(*ErrorResponse).
		sentForm()

	w, r := mockRequest(t, path, login, false)
	handler := LoginGoogle(dbMock, tokenAuth, gValidatorMock, expSessionLen, expTokenLen)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A malformed Google-Login didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A malformed Google-Login didn't return a valid errorResponse object") {
		assert.Equal(t, expResp, *resp, "A malformed Google-Login didn't return the proper error")
	}
	gValidatorMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
}

func TestTokenFailedLoginGoogle(t *testing.T) {
	path := "/auth/google"

	login := loginGoogleRequest{
		GoogleToken: "a.b.c",
	}

	gValidatorMock := gTokenValidatorMock{&mock.Mock{}}
	gValidatorMock.On("ValidateGToken", login.GoogleToken).
		Return(nil, errors.New("password wrong, should be xxxxx")).Once()

	dbMock := dBMock{&mock.Mock{}}

	expResp, expCode := ValidationFailedError(ErrLoginFailed).(*ErrorResponse).
		sentForm()

	w, r := mockRequest(t, path, login, false)
	handler := LoginGoogle(dbMock, tokenAuth, gValidatorMock, expSessionLen, expTokenLen)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A failed Google-Login didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A failed Google-Login didn't return a valid errorResponse object") {
		assert.Equal(t, expResp, *resp, "A failed Google-Login didn't return the proper error")
	}
	gValidatorMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
}

func TestFailedLoginGoogle(t *testing.T) {

	path := "/auth/google"

	login := loginGoogleRequest{
		GoogleToken: "a.b.c",
	}

	expGClaims := googlehelper.GoogleClaimsSchema{
		Email:         expID.Account,
		EmailVerified: true,
		FullName:      "Joko",
		AccountID:     expID.GAccount,
	}

	gValidatorMock := gTokenValidatorMock{&mock.Mock{}}

	gValidatorMock.On("ValidateGToken", login.GoogleToken).
		Return(&expGClaims, nil).Once()

	dbMock := dBMock{&mock.Mock{}}
	dbMock.On("LoginGoogle", expGClaims.Email, expGClaims.AccountID, expSessionLen).
		Return("", database.ErrAccountNotFound).Once()

	expResp, expCode := ValidationFailedError(ErrLoginFailed).(*ErrorResponse).
		sentForm()

	w, r := mockRequest(t, path, login, false)
	handler := LoginGoogle(dbMock, tokenAuth, gValidatorMock, expSessionLen, expTokenLen)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A failed Google-Login didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A failed Google-Login didn't return a valid errorResponse object") {
		assert.Equal(t, expResp, *resp, "A failed Google-Login didn't return the proper error")
	}
	gValidatorMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
}

func TestNotActivatedLoginGoogle(t *testing.T) {
	path := "/auth/google"

	login := loginGoogleRequest{
		GoogleToken: "a.b.c",
	}

	expGClaims := googlehelper.GoogleClaimsSchema{
		Email:         expID.Account,
		EmailVerified: true,
		FullName:      "Joko",
		AccountID:     expID.GAccount,
	}

	gValidatorMock := gTokenValidatorMock{&mock.Mock{}}
	gValidatorMock.On("ValidateGToken", login.GoogleToken).
		Return(&expGClaims, nil).Once()

	dbMock := dBMock{&mock.Mock{}}
	dbMock.On("LoginGoogle", expGClaims.Email, expGClaims.AccountID, expSessionLen).
		Return("", database.ErrAccountNotActive).Once()

	expResp, expCode := UnauthorizedRequestError(ErrLoginAccountNotActive).(*ErrorResponse).
		sentForm()

	w, r := mockRequest(t, path, login, false)
	handler := LoginGoogle(dbMock, tokenAuth, gValidatorMock, expSessionLen, expTokenLen)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A Google-Login on a not activated account didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A Google-Login on a not activated account didn't return a valid errorResponse object") {
		assert.Equal(t, expResp, *resp, "A Google-Login on a not activated account didn't return the proper error")
	}
	gValidatorMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
}
