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

func (db dBMock) RefreshActivation(ctx context.Context, email string, activationDuration time.Duration) (activationToken string, validUntil *time.Time, err error) {
	args := db.Called(email, activationDuration)
	if args.Get(1) == nil {
		return "", nil, args.Error(2)
	}
	return args.String(0), args.Get(1).(*time.Time), args.Error(2)
}
func TestSuccessfulResendActivationEmail(t *testing.T) {
	path := fmt.Sprintf("/auth/resend?email=%s", expID.Account)

	expDur := time.Minute * time.Duration(10)
	expTime := time.Now().Add(expDur)

	dbMock := dBMock{&mock.Mock{}}
	dbMock.On("RefreshActivation", expID.Account, expDur).
		Return(expID.AccountActivation, &expTime, nil)

	mailMock := activationMailDriverMock{&mock.Mock{}}
	mailMock.On("SendActivationEmail", expID.Account, expID.AccountActivation, expTime).
		Return(nil)

	expCode := http.StatusNoContent

	w, r := mockRequest(t, path, nil, false)
	handler := ResendActivationEmail(dbMock, mailMock, expDur)
	handler.ServeHTTP(w, r)

	resp := w.Body.String()

	assert.Equal(t, expCode, w.Code, "A failed Google-Login didn't return the proper response code")
	assert.Equal(t, "", resp, "A failed Google-Login didn't return the proper error")
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

	expResp, expCode := BadRequestError(errAccountActivationQueryMalformed).(*ErrorResponse).
		sentForm()

	resp := &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "A failed Google-Login didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A failed Google-Login didn't return a valid errorResponse object") {
		assert.Equal(t, expResp, *resp, "A failed Google-Login didn't return the proper error")
	}
	dbMock.AssertExpectations(t)
}
