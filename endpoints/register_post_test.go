package endpoints

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSuccessfulRegister(t *testing.T) {
	path := "/auth/register"

	reg := registerPostRequest{
		Email:    expID.Account,
		Password: "P4sswordꦏꦤ꧀",
		Name:     "Joko",
	}

	expTime := time.Now().Add(time.Minute * time.Duration(2))
	dbMock := &dBMock{}
	dbMock.On("Register", reg.Email, reg.Password, reg.Name).
		Return(expID.AccountActivation, &expTime, nil).Once()

	mailMock := &activationMailDriverMock{}
	mailMock.On("SendActivationEmail", reg.Email, expID.AccountActivation, expTime).
		Return(nil)

	expCode := http.StatusCreated

	r, w := mockRequest(t, path, reg, false)
	handler := RegisterPost(dbMock, tokenAuth, mailMock)
	handler.ServeHTTP(w, r)

	resp := &registerResponse{}
	assert.Equal(t, expCode, w.Code, "A successful Post-Register didn't return the proper response code")
	assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful Post-Register didn't return a valid registerResponse object")
	assert.Equal(t, reg.Email, resp.NewId, "A successful Post-Register didn't return a valid response")
	dbMock.AssertExpectations(t)
	mailMock.AssertExpectations(t)
}