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

	gValidatorMock := &gTokenValidatorMock{}
	gValidatorMock.On("ValidateGToken", login.GoogleToken).
		Return(&expGClaims, nil).Once()

	dbMock := &dBMock{}
	dbMock.On("LoginGoogle", expGClaims.Email, expGClaims.AccountID, expSessionLen).
		Return(expID.Session, nil).Once()

	expCode := http.StatusOK
	expClaims := sessiontoken.TokenClaimsSchema{
		Email:   expID.Account,
		Session: expID.Session,
	}

	r, w := mockRequest(t, path, login, false)
	handler := LoginGoogle(dbMock, tokenAuth, gValidatorMock, expSessionLen, expTokenLen)
	handler.ServeHTTP(w, r)

	resp := &tokenResponse{}
	assert.Equal(t, expCode, w.Code, "A successful Google-Login didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful Google-Login didn't return a valid tokenResponse object")
	token, err := jwtauth.VerifyToken(tokenAuth, resp.Token)
	assert.NoError(t, err, "A successful Google-Login didn't return a valid token")

	tokenMap, err := token.AsMap(context.Background())
	if assert.NoErrorf(t, err, "an error '%s' was not expected when getting returned token's schema") {
		var claims sessiontoken.TokenClaimsSchema
		err := claims.FromInterface(tokenMap)
		if assert.NoErrorf(t, err, "an error '%s' was not expected when getting returned token's schema") {
			assert.Equal(t, expClaims, claims, "A successful Post-Login didn't return expected token schema")
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

	gValidatorMock := &gTokenValidatorMock{}
	dbMock := &dBMock{}

	expResp, expCode := BadRequestError(ErrLoginGoogleMalformed).(*ErrorResponse).
		sentForm()

	r, w := mockRequest(t, path, login, false)
	handler := LoginGoogle(dbMock, tokenAuth, gValidatorMock, expSessionLen, expTokenLen)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A malformed Google-Login didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A malformed Google-Login didn't return a valid errorResponse object")
	assert.Equal(t, expResp, *resp, "A malformed Google-Login didn't return the proper error")
	gValidatorMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
}

func TestTokenFailedLoginGoogle(t *testing.T) {
	path := "/auth/google"

	login := loginGoogleRequest{
		GoogleToken: "a.b.c",
	}

	gValidatorMock := &gTokenValidatorMock{}
	gValidatorMock.On("ValidateGToken", login.GoogleToken).
		Return(nil, errors.New("password wrong, should be xxxxx")).Once()

	dbMock := &dBMock{}

	expResp, expCode := ValidationFailedError(ErrLoginFailed).(*ErrorResponse).
		sentForm()

	r, w := mockRequest(t, path, login, false)
	handler := LoginGoogle(dbMock, tokenAuth, gValidatorMock, expSessionLen, expTokenLen)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A failed Google-Login didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A failed Google-Login didn't return a valid errorResponse object")
	assert.Equal(t, expResp, *resp, "A failed Google-Login didn't return the proper error")
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

	gValidatorMock := &gTokenValidatorMock{}

	gValidatorMock.On("ValidateGToken", login.GoogleToken).
		Return(&expGClaims, nil).Once()

	dbMock := &dBMock{}
	dbMock.On("LoginGoogle", expGClaims.Email, expGClaims.AccountID, expSessionLen).
		Return("", database.ErrAccountNotFound).Once()

	expResp, expCode := ValidationFailedError(ErrLoginFailed).(*ErrorResponse).
		sentForm()

	r, w := mockRequest(t, path, login, false)
	handler := LoginGoogle(dbMock, tokenAuth, gValidatorMock, expSessionLen, expTokenLen)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A failed Google-Login didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A failed Google-Login didn't return a valid errorResponse object")
	assert.Equal(t, expResp, *resp, "A failed Google-Login didn't return the proper error")
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

	gValidatorMock := &gTokenValidatorMock{}
	gValidatorMock.On("ValidateGToken", login.GoogleToken).
		Return(&expGClaims, nil).Once()

	dbMock := &dBMock{}
	dbMock.On("LoginGoogle", expGClaims.Email, expGClaims.AccountID, expSessionLen).
		Return("", database.ErrAccountNotActive).Once()

	expResp, expCode := UnauthorizedRequestError(ErrLoginAccountNotActive).(*ErrorResponse).
		sentForm()

	r, w := mockRequest(t, path, login, false)
	handler := LoginGoogle(dbMock, tokenAuth, gValidatorMock, expSessionLen, expTokenLen)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}

	assert.Equal(t, expCode, w.Code, "A Google-Login on a not activated account didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A Google-Login on a not activated account didn't return a valid errorResponse object")
	assert.Equal(t, expResp, *resp, "A Google-Login on a not activated account didn't return the proper error")
	gValidatorMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
}
