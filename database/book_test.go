package database

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuccessfulSearchBooks(t *testing.T) {
	ctx := context.Background()

	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)
	defer d.Close()

	db := DBInstance{d}

	expBook := Book{
		ID:      uuid.MustParse("d42c75ef-5f76-4f26-8b8a-71fecd99b4f5"),
		Title:   "E",
		Author:  "MC2",
		Cover:   URLMustParse("https://example.com"),
		Readers: 10,
		IsFav:   true,
	}
	expBooks := []Book{
		expBook,
		expBook,
		expBook,
		expBook,
		expBook,
	}

	var rows = sqlmock.
		NewRows([]string{
			"b.id",
			"b.title",
			"b.cover_image",
			"b.author",
			"b.readers_count",
			"is_favorited"})
	for _, b := range expBooks {
		rows.AddRow(
			b.ID,
			b.Title,
			b.Cover.String(),
			b.Author,
			b.Readers,
			b.IsFav,
		)
	}

	expQuery := "potato"

	mock.ExpectPrepare("SELECT").
		ExpectQuery().
		WithArgs(expQuery).
		WillReturnRows(rows).
		RowsWillBeClosed()

	err = searchBooks.Prepare(ctx, d)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	books, err := db.SearchBooks(ctx, expQuery, expEmail)
	if assert.Nil(t, err, "unexpected error in a successful search books test") {
		assert.Equal(t, expBooks, books, "function should've returned a list of searched-for books")
	}
	assert.Nil(t, mock.ExpectationsWereMet())
}
