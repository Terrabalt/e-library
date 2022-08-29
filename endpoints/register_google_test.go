package endpoints

import (
	"encoding/json"
	"errors"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/googlehelper"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSuccessfulRegisterGoogle(t *testing.T) {
	path := "/auth/register/google"

	reg := registerGoogleRequest{
		GoogleToken: "a.b.c",
	}

	expGClaims := googlehelper.GoogleClaimsSchema{
		Email:         expID.Account,
		EmailVerified: true,
		FullName:      "Joko",
		AccountID:     expID.GAccount,
	}

	gValidatorMock := &gTokenValidatorMock{}
	gValidatorMock.On("ValidateGToken", reg.GoogleToken).
		Return(&expGClaims, nil).Once()

	expTime := time.Now().Add(time.Minute * time.Duration(2))
	dbMock := &dBMock{}
	dbMock.On("RegisterGoogle", expGClaims.Email, expGClaims.AccountID, expGClaims.FullName).
		Return(expID.AccountActivation, &expTime, nil).Once()

	mailMock := &activationMailDriverMock{}
	mailMock.On("SendActivationEmail", expGClaims.Email, expID.AccountActivation, expTime).
		Return(nil)

	expCode := http.StatusCreated

	r, w := mockRequest(t, path, reg, false)
	handler := RegisterGoogle(dbMock, tokenAuth, gValidatorMock, mailMock)
	handler.ServeHTTP(w, r)

	resp := &registerResponse{}
	assert.Equal(t, expCode, w.Code, "A malformed Google-Register didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A malformed Google-Register didn't return a valid registerResponse object")
	assert.Equal(t, expGClaims.Email, resp.NewId, "A malformed Google-Register didn't return a valid response")
	gValidatorMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
	mailMock.AssertExpectations(t)
}

func TestMalformedRegisterGoogle(t *testing.T) {
	path := "/auth/register/google"

	reg := registerGoogleRequest{
		GoogleToken: "",
	}

	gValidatorMock := &gTokenValidatorMock{}

	dbMock := &dBMock{}

	mailMock := &activationMailDriverMock{}

	expResp, expCode := BadRequestError(errRegisterGoogleMalformed).(*ErrorResponse).sentForm()

	r, w := mockRequest(t, path, reg, false)
	handler := RegisterGoogle(dbMock, tokenAuth, gValidatorMock, mailMock)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "A successful Google-Register didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful Google-Register didn't return a valid registerResponse object")
	assert.Equal(t, expResp, *resp, "A successful Google-Register didn't return a valid response")
	gValidatorMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
	mailMock.AssertExpectations(t)
}

func TestNonValidatedRegisterGoogle(t *testing.T) {
	path := "/auth/register/google"

	reg := registerGoogleRequest{
		GoogleToken: "a.b.c",
	}

	gValidatorMock := &gTokenValidatorMock{}
	gValidatorMock.On("ValidateGToken", reg.GoogleToken).
		Return(nil, errors.New("")).Once()

	dbMock := &dBMock{}

	mailMock := &activationMailDriverMock{}

	expResp, expCode := ValidationFailedError(errGoogleTokenFailed).(*ErrorResponse).sentForm()

	r, w := mockRequest(t, path, reg, false)
	handler := RegisterGoogle(dbMock, tokenAuth, gValidatorMock, mailMock)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "A token-failed Google-Register didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A token-failed Google-Register didn't return a valid registerResponse object")
	assert.Equal(t, expResp, *resp, "A token-failed Google-Register didn't return a valid response")
	gValidatorMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
	mailMock.AssertExpectations(t)
}

func TestAlreadyRegisteredGoogle(t *testing.T) {
	path := "/auth/register/google"

	reg := registerGoogleRequest{
		GoogleToken: "a.b.c",
	}

	expGClaims := googlehelper.GoogleClaimsSchema{
		Email:         expID.Account,
		EmailVerified: true,
		FullName:      "Joko",
		AccountID:     expID.GAccount,
	}

	gValidatorMock := &gTokenValidatorMock{}
	gValidatorMock.On("ValidateGToken", reg.GoogleToken).
		Return(&expGClaims, nil).Once()

	dbMock := &dBMock{}
	dbMock.On("RegisterGoogle", expGClaims.Email, expGClaims.AccountID, expGClaims.FullName).
		Return("", nil, database.ErrAccountExisted).Once()

	mailMock := &activationMailDriverMock{}

	expResp, expCode := RequestConflictError(errAccountAlreadyRegistered).(*ErrorResponse).sentForm()

	r, w := mockRequest(t, path, reg, false)
	handler := RegisterGoogle(dbMock, tokenAuth, gValidatorMock, mailMock)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "An already-registered Google-Register didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "An already-registered Google-Register didn't return a valid registerResponse object")
	assert.Equal(t, expResp, *resp, "An already-registered Google-Register didn't return a valid response")
	gValidatorMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
	mailMock.AssertExpectations(t)
}
