package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const expSessionToken = "abcdefgh"

func TestSuccessfulCheckSession(t *testing.T) {
	ctx := context.Background()

	expOut := true
	expTime := time.Now()

	var rows = sqlmock.
		NewRows([]string{"session_valid"}).
		AddRow(expOut)

	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer d.Close()

	db := DBInstance{d}

	mock.ExpectPrepare("SELECT").
		ExpectQuery().
		WithArgs(expEmail, expSessionToken, expTime).
		WillReturnRows(rows).
		RowsWillBeClosed()

	err = isValidSessionQuery.Prepare(ctx, d)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	out, err := db.CheckSession(ctx, expEmail, expSessionToken, expTime)
	if assert.Nil(t, err, "unexpected error in a successful get new books test") {
		assert.Equal(t, expOut, out, "function should've returned a list of new books")
	}
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestFailedCheckSession(t *testing.T) {
	ctx := context.Background()

	expOut := false
	expTime := time.Now()

	var rows = sqlmock.
		NewRows([]string{"session_valid"}).
		AddRow(expOut)

	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer d.Close()

	db := DBInstance{d}

	test := mock.ExpectPrepare("SELECT")
	test.ExpectQuery().
		WithArgs(expEmail, expSessionToken, expTime).
		WillReturnRows(rows).
		RowsWillBeClosed()

	err = isValidSessionQuery.Prepare(ctx, d)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	out, err := db.CheckSession(ctx, expEmail, expSessionToken, expTime)
	if assert.Equal(t, ErrSessionInvalid, err, "unexpected error in a successful get new books test") {
		assert.Equal(t, expOut, out, "function should've returned a list of new books")
	}
	assert.Nil(t, mock.ExpectationsWereMet())

	rows = sqlmock.
		NewRows([]string{"session_valid"})

	test.ExpectQuery().
		WithArgs(expEmail, expSessionToken, expTime).
		WillReturnRows(rows).
		RowsWillBeClosed()

	out, err = db.CheckSession(ctx, expEmail, expSessionToken, expTime)
	if assert.Equal(t, ErrSessionInvalid, err, "unexpected error in a successful get new books test") {
		assert.Equal(t, expOut, out, "function should've returned a list of new books")
	}
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestErrorCheckSession(t *testing.T) {
	ctx := context.Background()

	expOut := false
	expTime := time.Now()

	var row = errors.New("a")

	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer d.Close()

	db := DBInstance{d}

	mock.ExpectPrepare("SELECT").
		ExpectQuery().
		WithArgs(expEmail, expSessionToken, expTime).
		WillReturnError(row)

	err = isValidSessionQuery.Prepare(ctx, d)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	out, err := db.CheckSession(ctx, expEmail, expSessionToken, expTime)
	if assert.Equal(t, row, err, "unexpected error in a successful get new books test") {
		assert.Equal(t, expOut, out, "function should've returned a list of new books")
	}
	assert.Nil(t, mock.ExpectationsWereMet())
}
