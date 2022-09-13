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

func (db dBMock) GetNewBooksPaginated(ctx context.Context, limit int, offset int, accountID string) ([]database.Book, error) {
	args := db.Called(limit, offset, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]database.Book), args.Error(1)
}

func TestSuccessfulHomepageListNewBooks(t *testing.T) {
	path := "/books?criteria=newHomepage"

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
	dbMock.On("GetNewBooksPaginated", 8, 0, expID.Account).
		Return(expDBBooks, nil).Once()

	expCode := http.StatusOK

	w, r := mockRequest(t, path, nil, true)
	handler := ListBooks(dbMock)
	handler.ServeHTTP(w, r)

	expResp := BooksFromDatabase(expDBBooks)

	resp := &BooksResponse{}
	assert.Equal(t, expCode, w.Code, "A successful Homepage-New-Books-List didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful Homepage-New-Books-List didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "A successful Homepage-New-Books-List didn't return a valid response")
	}
	dbMock.AssertExpectations(t)

	expDBBooks = []database.Book{}

	dbMock.On("GetNewBooksPaginated", 8, 0, expID.Account).
		Return(expDBBooks, nil).Once()

	w, r = mockRequest(t, path, nil, true)
	handler = ListBooks(dbMock)
	handler.ServeHTTP(w, r)

	expResp = BooksFromDatabase(expDBBooks)

	resp = &BooksResponse{}
	assert.Equal(t, expCode, w.Code, "An empty Homepage-New-Books-List didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A successful Homepage-New-Books-List didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "An empty Homepage-New-Books-List didn't return a valid response")
	}
	dbMock.AssertExpectations(t)
}

func TestFailedHomepageListNewBooks(t *testing.T) {
	path := "/books?criteria=newHomepage"

	dbMock := dBMock{&mock.Mock{}}

	w, r := mockRequest(t, path, nil, false)
	handler := ListBooks(dbMock)
	handler.ServeHTTP(w, r)

	expResp, expCode := InternalServerError().(*ErrorResponse).sentForm()

	resp := &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "A failed Homepage-New-Books-List didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A failed Homepage-New-Books-List didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "A failed Homepage-New-Books-List didn't return a valid response")
	}
	dbMock.AssertExpectations(t)

	dbMock.On("GetNewBooksPaginated", 8, 0, expID.Account).
		Return(nil, sql.ErrConnDone).Once()

	w, r = mockRequest(t, path, nil, true)
	handler = ListBooks(dbMock)
	handler.ServeHTTP(w, r)

	expResp, expCode = InternalServerError().(*ErrorResponse).sentForm()

	resp = &ErrorResponse{}
	assert.Equal(t, expCode, w.Code, "A errored Homepage-New-Books-List didn't return the proper response code")
	if assert.Nil(t, json.NewDecoder(w.Body).Decode(resp), "A errored Homepage-New-Books-List didn't return a valid registerResponse object") {
		assert.Equal(t, expResp, *resp, "A errored Homepage-New-Books-List didn't return a valid response")
	}
	dbMock.AssertExpectations(t)
}
