package endpoints

import (
	"context"
	"database/sql"
	"encoding/json"
	"ic-rhadi/e_library/database"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (db dBMock) SearchBooks(ctx context.Context, query string, accountID string) ([]database.Book, error) {
	args := db.Called(query, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]database.Book), args.Error(1)
}

func TestSuccessfulSearchBooks(t *testing.T) {
	expQuery := "potato"
	path := "/books/search?query=" + expQuery

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
	dbMock.On("SearchBooks", expQuery, expID.Account).
		Return(expDBBooks, nil).Once()

	expCode := http.StatusOK

	w, r := mockRequest(t, path, nil, true)
	handler := SearchBooks(dbMock)
	handler.ServeHTTP(w, r)

	expResp := BooksFromDatabase(expDBBooks)

	resp := &BooksResponse{}
	assert.Equal(t, expCode, w.Code, "A successful Books-Search didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful Books-Search didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "A successful Books-Search didn't return a valid response")
	}
	dbMock.AssertExpectations(t)

	expDBBooks = []database.Book{}

	dbMock.On("SearchBooks", expQuery, expID.Account).
		Return(expDBBooks, nil).Once()

	w, r = mockRequest(t, path, nil, true)
	handler = SearchBooks(dbMock)
	handler.ServeHTTP(w, r)

	expResp = BooksFromDatabase(expDBBooks)

	resp = &BooksResponse{}
	assert.Equal(t, expCode, w.Code, "An empty Books-Search didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful Books-Search didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "An empty Books-Search didn't return a valid response")
	}
	dbMock.AssertExpectations(t)
}

func TestFailedSearchBooks(t *testing.T) {
	expQuery := "potato"
	path := "/books/search?query=" + expQuery

	dbMock := dBMock{&mock.Mock{}}

	w, r := mockRequest(t, path, nil, false)
	handler := SearchBooks(dbMock)
	handler.ServeHTTP(w, r)

	expResp, expCode := InternalServerError().(*ErrorResponse).sentForm()

	resp := &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "A failed Books-Search didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A failed Books-Search didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "A failed Books-Search didn't return a valid response")
	}
	dbMock.AssertExpectations(t)

	dbMock = dBMock{&mock.Mock{}}
	dbMock.On("SearchBooks", expQuery, expID.Account).
		Return(nil, sql.ErrConnDone).Once()

	w, r = mockRequest(t, path, nil, true)
	handler = SearchBooks(dbMock)
	handler.ServeHTTP(w, r)

	resp = &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "A errored Books-Search didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A errored Books-Search didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "A errored Books-Search didn't return a valid response")
	}
	dbMock.AssertExpectations(t)
}
