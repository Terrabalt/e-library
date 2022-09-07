package endpoints

import (
	"encoding/json"
	"ic-rhadi/e_library/database"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSuccessfulRegister(t *testing.T) {
	path := "/auth/register"

	reg := registerPostRequest{
		Email:    expID.Account,
		Password: "P4sswordꦏꦤ꧀",
		Name:     "Joko",
	}

	expDur := time.Minute * time.Duration(2)
	expTime := time.Now().Add(expDur)
	dbMock := dBMock{&mock.Mock{}}
	dbMock.On("Register", reg.Email, reg.Password, reg.Name, expDur).
		Return(expID.AccountActivation, &expTime, nil).Once()

	mailMock := activationMailDriverMock{&mock.Mock{}}
	mailMock.On("SendActivationEmail", reg.Email, expID.AccountActivation, expTime).
		Return(nil)

	expCode := http.StatusCreated

	w, r := mockRequest(t, path, reg, false)
	handler := RegisterPost(dbMock, tokenAuth, mailMock, expDur)
	handler.ServeHTTP(w, r)

	resp := &registerResponse{}
	assert.Equal(t, expCode, w.Code, "A successful Post-Register didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful Post-Register didn't return a valid registerResponse object")
	assert.Equal(t, reg.Email, resp.NewID, "A successful Post-Register didn't return a valid response")
	dbMock.AssertExpectations(t)
	mailMock.AssertExpectations(t)
}

func TestMalformedRegistered(t *testing.T) {
	path := "/auth/register"

	reg := registerPostRequest{
		Email:    "username",
		Password: "P4sswordꦏꦤ꧀",
		Name:     "Joko",
	}

	expDur := time.Minute * time.Duration(2)
	dbMock := dBMock{&mock.Mock{}}

	mailMock := activationMailDriverMock{&mock.Mock{}}

	expResp, expCode := BadRequestError(errEmailMalformed).(*ErrorResponse).sentForm()

	w, r := mockRequest(t, path, reg, false)
	handler := RegisterPost(dbMock, tokenAuth, mailMock, expDur)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "An already-registered Post-Register didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "An already-registered Post-Register didn't return a valid registerResponse object")
	assert.Equal(t, expResp, *resp, "An already-registered Post-Register didn't return a valid response")

	reg = registerPostRequest{
		Email:    expID.Account,
		Password: "Passwordꦏꦤ꧀",
		Name:     "Joko",
	}

	expResp, expCode = BadRequestError(errPasswordDontHaveNumber).(*ErrorResponse).sentForm()

	w, r = mockRequest(t, path, reg, false)
	handler = RegisterPost(dbMock, tokenAuth, mailMock, expDur)
	handler.ServeHTTP(w, r)

	resp = &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "An already-registered Post-Register didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "An already-registered Post-Register didn't return a valid registerResponse object")
	assert.Equal(t, expResp, *resp, "An already-registered Post-Register didn't return a valid response")

	reg = registerPostRequest{
		Email:    expID.Account,
		Password: "p4sswordꦏꦤ꧀",
		Name:     "Joko",
	}

	expResp, expCode = BadRequestError(errPasswordDontHaveUppercase).(*ErrorResponse).sentForm()

	w, r = mockRequest(t, path, reg, false)
	handler = RegisterPost(dbMock, tokenAuth, mailMock, expDur)
	handler.ServeHTTP(w, r)

	resp = &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "An already-registered Post-Register didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "An already-registered Post-Register didn't return a valid registerResponse object")
	assert.Equal(t, expResp, *resp, "An already-registered Post-Register didn't return a valid response")

	reg = registerPostRequest{
		Email:    expID.Account,
		Password: "P4ssword",
		Name:     "Joko",
	}

	expResp, expCode = BadRequestError(errPasswordDontHaveSpecials).(*ErrorResponse).sentForm()

	w, r = mockRequest(t, path, reg, false)
	handler = RegisterPost(dbMock, tokenAuth, mailMock, expDur)
	handler.ServeHTTP(w, r)

	resp = &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "An already-registered Post-Register didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "An already-registered Post-Register didn't return a valid registerResponse object")
	assert.Equal(t, expResp, *resp, "An already-registered Post-Register didn't return a valid response")

	reg = registerPostRequest{
		Email:    expID.Account,
		Password: "P4ss",
		Name:     "Joko",
	}

	expResp, expCode = BadRequestError(errPasswordTooShort).(*ErrorResponse).sentForm()

	w, r = mockRequest(t, path, reg, false)
	handler = RegisterPost(dbMock, tokenAuth, mailMock, expDur)
	handler.ServeHTTP(w, r)

	resp = &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "An already-registered Post-Register didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "An already-registered Post-Register didn't return a valid registerResponse object")
	assert.Equal(t, expResp, *resp, "An already-registered Post-Register didn't return a valid response")

	reg = registerPostRequest{
		Email:    expID.Account,
		Password: "P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀P4sswordꦏꦤ꧀",
		Name:     "Joko",
	}

	expResp, expCode = BadRequestError(errPasswordTooLong).(*ErrorResponse).sentForm()

	w, r = mockRequest(t, path, reg, false)
	handler = RegisterPost(dbMock, tokenAuth, mailMock, expDur)
	handler.ServeHTTP(w, r)

	resp = &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "An already-registered Post-Register didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "An already-registered Post-Register didn't return a valid registerResponse object")
	assert.Equal(t, expResp, *resp, "An already-registered Post-Register didn't return a valid response")
}

func TestAlreadyRegistered(t *testing.T) {
	path := "/auth/register"

	reg := registerPostRequest{
		Email:    expID.Account,
		Password: "P4sswordꦏꦤ꧀",
		Name:     "Joko",
	}

	expDur := time.Minute * time.Duration(2)
	dbMock := dBMock{&mock.Mock{}}
	dbMock.On("Register", reg.Email, reg.Password, reg.Name).
		Return("", nil, database.ErrAccountExisted).Once()

	mailMock := activationMailDriverMock{&mock.Mock{}}

	expResp, expCode := RequestConflictError(errAccountAlreadyRegistered).(*ErrorResponse).sentForm()

	w, r := mockRequest(t, path, reg, false)
	handler := RegisterPost(dbMock, tokenAuth, mailMock, expDur)
	handler.ServeHTTP(w, r)

	resp := &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "An already-registered Post-Register didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "An already-registered Post-Register didn't return a valid registerResponse object")
	assert.Equal(t, expResp, *resp, "An already-registered Post-Register didn't return a valid response")
	dbMock.AssertExpectations(t)
	mailMock.AssertExpectations(t)
}
