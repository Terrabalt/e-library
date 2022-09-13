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

func (db dBMock) GetPopularBooksPaginated(ctx context.Context, limit int, offset int, accountID string) ([]database.Book, error) {
	args := db.Called(limit, offset, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]database.Book), args.Error(1)
}

func TestSuccessfulHomepageListPopularBooks(t *testing.T) {
	path := "books/popular/homepage"

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
	dbMock.On("GetPopularBooksPaginated", 8, 0, expID.Account).
		Return(expDBBooks, nil).Once()

	expCode := http.StatusOK

	w, r := mockRequest(t, path, nil, true)
	handler := homepageListPopularBooks(dbMock)
	handler.ServeHTTP(w, r)

	expResp := BooksFromDatabase(expDBBooks)

	resp := &BooksResponse{}
	assert.Equal(t, expCode, w.Code, "A successful Homepage-Popular-Books-List didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful Homepage-Popular-Books-List didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "A successful Homepage-Popular-Books-List didn't return a valid response")
	}
	dbMock.AssertExpectations(t)

	expDBBooks = []database.Book{}

	dbMock.On("GetPopularBooksPaginated", 8, 0, expID.Account).
		Return(expDBBooks, nil).Once()

	w, r = mockRequest(t, path, nil, true)
	handler = homepageListPopularBooks(dbMock)
	handler.ServeHTTP(w, r)

	expResp = BooksFromDatabase(expDBBooks)

	resp = &BooksResponse{}
	assert.Equal(t, expCode, w.Code, "A successful Homepage-Popular-Books-List didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful Homepage-Popular-Books-List didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "A successful Homepage-Popular-Books-List didn't return a valid response")
	}
	dbMock.AssertExpectations(t)
}

func TestFailedHomepageListPopularBooks(t *testing.T) {
	path := "books/popular/homepage"

	dbMock := dBMock{&mock.Mock{}}

	w, r := mockRequest(t, path, nil, false)
	handler := homepageListPopularBooks(dbMock)
	handler.ServeHTTP(w, r)

	expResp, expCode := InternalServerError().(*ErrorResponse).sentForm()

	resp := &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "A failed Homepage-Popular-Books-List didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A failed Homepage-Popular-Books-List didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "A failed Homepage-Popular-Books-List didn't return a valid response")
	}
	dbMock.AssertExpectations(t)

	dbMock.On("GetPopularBooksPaginated", 8, 0, expID.Account).
		Return(nil, sql.ErrConnDone).Once()

	w, r = mockRequest(t, path, nil, true)
	handler = homepageListPopularBooks(dbMock)
	handler.ServeHTTP(w, r)

	resp = &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "A errored Homepage-Popular-Books-List didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A errored Homepage-Popular-Books-List didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "A errored Homepage-Popular-Books-List didn't return a valid response")
	}
	dbMock.AssertExpectations(t)
}
