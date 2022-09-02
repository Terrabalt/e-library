package endpoints

import (
	"database/sql"
	"encoding/json"
	"ic-rhadi/e_library/database"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSuccessfulListMoreNewBooks(t *testing.T) {
	path := "/books/new/more"

	expDBBook := database.Book{
		ID:      uuid.New(),
		Title:   "aaaa",
		Author:  "ae",
		Readers: 10,
	}
	expDBBooks := []database.Book{
		expDBBook,
		expDBBook,
		expDBBook,
		expDBBook,
		expDBBook,
	}

	dbMock := dBMock{&mock.Mock{}}
	dbMock.On("GetNewBooks", 0, 0, expID.Account).
		Return(expDBBooks, nil).Once()

	expCode := http.StatusOK

	w, r := mockRequest(t, path, nil, true)
	handler := ListMoreNewBooks(dbMock)
	handler.ServeHTTP(w, r)

	expResp := BooksFromDatabase(expDBBooks)

	resp := &BooksResponse{}
	assert.Equal(t, expCode, w.Code, "A successful New-Books-List didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful New-Books-List didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "A successful New-Books-List didn't return a valid response")
	}
	dbMock.AssertExpectations(t)

	expDBBooks = []database.Book{}

	dbMock.On("GetNewBooks", 0, 0, expID.Account).
		Return(expDBBooks, nil).Once()

	w, r = mockRequest(t, path, nil, true)
	handler = ListMoreNewBooks(dbMock)
	handler.ServeHTTP(w, r)

	expResp = BooksFromDatabase(expDBBooks)

	resp = &BooksResponse{}
	assert.Equal(t, expCode, w.Code, "A successful New-Books-List didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful New-Books-List didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "A successful New-Books-List didn't return a valid response")
	}
	dbMock.AssertExpectations(t)
}

func TestFailedListMoreNewBooks(t *testing.T) {
	path := "/books/new/more"

	dbMock := dBMock{&mock.Mock{}}

	w, r := mockRequest(t, path, nil, false)
	handler := ListMoreNewBooks(dbMock)
	handler.ServeHTTP(w, r)

	expResp, expCode := InternalServerError().(*ErrorResponse).sentForm()

	resp := &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "A failed New-Books-List didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A failed New-Books-List didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "A failed New-Books-List didn't return a valid response")
	}
	dbMock.AssertExpectations(t)

	dbMock.On("GetNewBooks", 0, 0, expID.Account).
		Return(nil, sql.ErrConnDone).Once()

	w, r = mockRequest(t, path, nil, true)
	handler = ListMoreNewBooks(dbMock)
	handler.ServeHTTP(w, r)

	resp = &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "A errored New-Books-List didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A errored New-Books-List didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "A errored New-Books-List didn't return a valid response")
	}
	dbMock.AssertExpectations(t)
}
