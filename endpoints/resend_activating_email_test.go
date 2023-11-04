package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (db dBMock) GetActivationData(ctx context.Context, email string) (activated bool, activationToken string, expiresIn *time.Time, err error) {
	args := db.Called(email)

	return args.Bool(0), args.String(1), args.Get(2).(*time.Time), args.Error(3)
}
func (db dBMock) RefreshActivation(ctx context.Context, email string, activationToken string, validUntil time.Time) error {
	args := db.Called(email, activationToken, validUntil)

	return args.Error(0)
}

func TestSuccessfulResendActivationEmail(t *testing.T) {
	path := fmt.Sprintf("/auth/resend?email=%s", expID.Account)

	expDur := time.Minute * time.Duration(10)
	expTime := time.Now().Add(expDur)

	var expNewActivation string
	dbMock := dBMock{&mock.Mock{}}
	dbMock.On("RefreshActivation", expID.Account, mock.MatchedBy(func(req string) bool { expNewActivation = req; return true }), mock.MatchedBy(func(req time.Time) bool { return req.After(expTime) })).
		Return(nil)

	mailMock := activationMailDriverMock{&mock.Mock{}}
	mailMock.On("SendActivationEmail", expID.Account, expNewActivation, mock.MatchedBy(func(req time.Time) bool { return req.After(expTime) })).
		Return(nil)

	w, r := mockRequest(t, path, nil, false)
	handler := ResendActivationEmail(dbMock, mailMock, expDur)
	handler.ServeHTTP(w, r)

	expResp := resendResponse{"resend succesful"}
	expCode := http.StatusOK

	resp := &resendResponse{}
	assert.Equal(t, expCode, w.Code, "A successful activation email resend didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful activation email resend didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "A successful activation email resend didn't return a valid response")
	}
	dbMock.AssertExpectations(t)
	mailMock.AssertExpectations(t)
}

func TestMalformedResendActivationEmail(t *testing.T) {
	path := fmt.Sprintf("/auth/resend?emai=%s", expID.Account)

	expDur := time.Minute * time.Duration(10)

	dbMock := dBMock{&mock.Mock{}}
	mailMock := activationMailDriverMock{&mock.Mock{}}

	w, r := mockRequest(t, path, nil, false)
	handler := ResendActivationEmail(dbMock, mailMock, expDur)
	handler.ServeHTTP(w, r)

	expResp, expCode := BadRequestError(errResendActivationEmailMalformed).(*ErrorResponse).
		sentForm()

	resp := &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "A malformed activation email resend didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A malformed activation email resend didn't return a valid errorResponse object") {
		assert.Equal(t, expResp, *resp, "A malformed activation email resend didn't return the proper error")
	}
	dbMock.AssertExpectations(t)
	mailMock.AssertExpectations(t)
}
