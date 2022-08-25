package endpoints

import (
	"encoding/json"
	"ic-rhadi/e_library/database"
	"net/http"
	"testing"

	"github.com/go-chi/jwtauth"
	"github.com/stretchr/testify/assert"
)

func TestSuccessfulLoginPost(t *testing.T) {
	path := "/auth/login"

	login := loginPostRequest{
		Email:    expId.Account,
		Password: "Password",
	}

	dbMock := &dBMock{}
	dbMock.On("Login", login.Email, login.Password).
		Return(expId.Session, nil).Once()

	expCode := http.StatusOK

	r, w := mockRequest(t, path, login, false)
	handler := LoginPostEndpoint(dbMock, tokenAuth)
	handler.ServeHTTP(w, r)

	resp := &tokenResponse{}
	assert.Equal(t, expCode, w.Code, "A successful Post-Login didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful Post-Login didn't return a valid tokenResponse object")
	token, err := jwtauth.VerifyToken(tokenAuth, resp.Token)
	assert.NoError(t, err, "A successful Post-Login didn't return a valid token")

	email, _ := token.Get("sub")
	session, _ := token.Get("session")
	assert.Equal(t, expId.Account, email, "A successful Post-Login didn't return expected email")
	assert.Equal(t, expId.Session, session, "A successful Post-Login didn't return expected session id")
	dbMock.AssertExpectations(t)
}

func TestMalformedLoginPost(t *testing.T) {
	path := "/auth/login"

	login := loginPostRequest{
		Email:    "",
		Password: "Password",
	}

	dbMock := &dBMock{}

	expResp, expCode := BadRequestError(ErrLoginPostMalformed).(*ErrorResponse).
		sentForm()

	r, w := mockRequest(t, path, login, false)
	handler := LoginPostEndpoint(dbMock, tokenAuth)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A malformed Post-Login didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A malformed Post-Login didn't return a valid errorResponse object")
	assert.Equal(t, expResp, *resp, "A malformed Post-Login didn't return the proper error")
	dbMock.AssertExpectations(t)
}

func TestFailedLoginPost(t *testing.T) {
	path := "/auth/login"

	login := loginPostRequest{
		Email:    expId.Account,
		Password: "Passwor",
	}

	dbMock := &dBMock{}
	dbMock.On("Login", login.Email, login.Password).
		Return("", database.ErrAccountNotFound).Once()

	expResp, expCode := ValidationFailedError(ErrLoginFailed).(*ErrorResponse).
		sentForm()

	r, w := mockRequest(t, path, login, false)
	handler := LoginPostEndpoint(dbMock, tokenAuth)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A failed Post-Login didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A failed Post-Login didn't return a valid errorResponse object")
	assert.Equal(t, expResp, *resp, "A failed Post-Login didn't return the proper error")
	dbMock.AssertExpectations(t)
}

func TestNotActivatedLoginPost(t *testing.T) {
	path := "/auth/login"

	login := loginPostRequest{
		Email:    expId.Account,
		Password: "Password",
	}

	dbMock := &dBMock{}
	dbMock.On("Login", login.Email, login.Password).
		Return("", database.ErrAccountNotActive).Once()

	expResp, expCode := UnauthorizedRequestError(ErrLoginAccountNotActive).(*ErrorResponse).
		sentForm()

	r, w := mockRequest(t, path, login, false)
	handler := LoginPostEndpoint(dbMock, tokenAuth)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A Post-Login on a not activated account didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A Post-Login on a not activated account didn't return a valid errorResponse object")
	assert.Equal(t, expResp, *resp, "A Post-Login on a not activated account didn't return the proper error")
	dbMock.AssertExpectations(t)
}
