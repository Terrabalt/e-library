package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (db dBMock) ActivateAccount(ctx context.Context, email string) error {
	args := db.Called(email)
	return args.Error(0)
}

func TestSuccessfulActivateAccount(t *testing.T) {
	path := fmt.Sprintf("/auth/activate?email=%s&token=%s", expID.Account, expID.AccountActivation)

	dbMock := dBMock{&mock.Mock{}}
	dbMock.On("ActivateAccount", expID.Account, expID.AccountActivation).
		Return(nil)

	w, r := mockRequest(t, path, nil, false)
	handler := ActivateAccount(dbMock)
	handler.ServeHTTP(w, r)

	expResp := activatedResponse{
		Message: "account activated",
		Email:   expID.Account,
	}
	expCode := http.StatusOK

	resp := &activatedResponse{}
	assert.Equal(t, expCode, w.Code, "A successful account activation didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful account activation didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "A successful account activation didn't return a valid response")
	}
	dbMock.AssertExpectations(t)
}

func TestMalformedActivateAccount(t *testing.T) {
	path := fmt.Sprintf("/auth/activate?emai=%s&token=%s", expID.Account, expID.AccountActivation)

	dbMock := dBMock{&mock.Mock{}}

	w, r := mockRequest(t, path, nil, false)
	handler := ActivateAccount(dbMock)
	handler.ServeHTTP(w, r)

	expResp, expCode := BadRequestError(errAccountActivationQueryMalformed).(*ErrorResponse).
		sentForm()

	resp := &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "A failed account activation didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A failed account activation didn't return a valid errorResponse object") {
		assert.Equal(t, expResp, *resp, "A failed account activation didn't return the proper error")
	}
	dbMock.AssertExpectations(t)
}
