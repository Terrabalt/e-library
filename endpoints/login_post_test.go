package endpoints

import (
	"context"
	"encoding/json"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/sessiontoken"
	"net/http"
	"testing"

	"github.com/go-chi/jwtauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSuccessfulLoginPost(t *testing.T) {
	path := "/auth/login"

	login := loginPostRequest{
		Email:    expID.Account,
		Password: "Password",
	}

	dbMock := dBMock{&mock.Mock{}}
	dbMock.On("Login", login.Email, login.Password, expSessionLen).
		Return(expID.Session, nil).Once()

	expCode := http.StatusOK
	expClaims := sessiontoken.TokenClaimsSchema{
		Email:   expID.Account,
		Session: expID.Session,
	}

	w, r := mockRequest(t, path, login, false)
	handler := LoginPost(dbMock, tokenAuth, expSessionLen, expTokenLen)
	handler.ServeHTTP(w, r)

	resp := &tokenResponse{}
	assert.Equal(t, expCode, w.Code, "A successful Post-Login didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful Post-Login didn't return a valid tokenResponse object") {
		token, err := jwtauth.VerifyToken(tokenAuth, resp.Token)
		assert.NoError(t, err, "A successful Post-Login didn't return a valid token")

		tokenMap, err := token.AsMap(context.Background())
		if assert.NoErrorf(t, err, "an error '%s' was not expected when getting returned token's schema") {
			var claims sessiontoken.TokenClaimsSchema
			err := claims.FromInterface(tokenMap)
			if assert.NoErrorf(t, err, "an error '%s' was not expected when getting returned token's schema") {
				assert.Equal(t, expClaims, claims, "A successful Post-Login didn't return expected token schema")
			}
		}
	}
	dbMock.AssertExpectations(t)
}

func TestMalformedLoginPost(t *testing.T) {
	path := "/auth/login"

	login := loginPostRequest{
		Email:    "",
		Password: "Password",
	}

	dbMock := dBMock{&mock.Mock{}}

	expResp, expCode := BadRequestError(ErrLoginPostMalformed).(*ErrorResponse).
		sentForm()

	w, r := mockRequest(t, path, login, false)
	handler := LoginPost(dbMock, tokenAuth, expSessionLen, expTokenLen)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A malformed Post-Login didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A malformed Post-Login didn't return a valid errorResponse object") {
		assert.Equal(t, expResp, *resp, "A malformed Post-Login didn't return the proper error")
	}
	dbMock.AssertExpectations(t)
}

func TestFailedLoginPost(t *testing.T) {
	path := "/auth/login"

	login := loginPostRequest{
		Email:    expID.Account,
		Password: "Passwor",
	}

	dbMock := dBMock{&mock.Mock{}}
	dbMock.On("Login", login.Email, login.Password, expSessionLen).
		Return("", database.ErrAccountNotFound).Once()

	expResp, expCode := ValidationFailedError(ErrLoginFailed).(*ErrorResponse).
		sentForm()

	w, r := mockRequest(t, path, login, false)
	handler := LoginPost(dbMock, tokenAuth, expSessionLen, expTokenLen)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A failed Post-Login didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A failed Post-Login didn't return a valid errorResponse object") {
		assert.Equal(t, expResp, *resp, "A failed Post-Login didn't return the proper error")
	}
	dbMock.AssertExpectations(t)
}

func TestNotActivatedLoginPost(t *testing.T) {
	path := "/auth/login"

	login := loginPostRequest{
		Email:    expID.Account,
		Password: "Password",
	}

	dbMock := dBMock{&mock.Mock{}}
	dbMock.On("Login", login.Email, login.Password, expSessionLen).
		Return("", database.ErrAccountNotActive).Once()

	expResp, expCode := UnauthorizedRequestError(ErrLoginAccountNotActive).(*ErrorResponse).
		sentForm()

	w, r := mockRequest(t, path, login, false)
	handler := LoginPost(dbMock, tokenAuth, expSessionLen, expTokenLen)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A Post-Login on a not activated account didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A Post-Login on a not activated account didn't return a valid errorResponse object") {
		assert.Equal(t, expResp, *resp, "A Post-Login on a not activated account didn't return the proper error")
	}
	dbMock.AssertExpectations(t)
}
