package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type testUUID struct {
	uuid string
}

func (tu *testUUID) Match(value driver.Value) bool {
	switch v := value.(type) {
	case []byte:
		uuid, err := uuid.ParseBytes(v)
		if err != nil {
			return false
		}
		tu.uuid = uuid.String()
		return true
	case string:
		uuid, err := uuid.Parse(v)
		if err != nil {
			return false
		}
		tu.uuid = uuid.String()
		return true
	default:
		return false
	}
}

const expEmail = "a@b.c"
const expPassword = "password"

func TestSuccessfulLogin(t *testing.T) {
	ctx := context.Background()

	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer d.Close()

	db := DBInstance{d}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(expPassword), 4)
	require.NoErrorf(t, err, "an error '%s' was not expected when creating a mock hashed password", err)

	var rowsPost = sqlmock.
		NewRows([]string{"password", "g_id", "activated"}).
		AddRow(hashPass, nil, true)

	test1 := mock.ExpectPrepare("SELECT")
	test2 := mock.ExpectPrepare("INSERT")
	mock.ExpectBegin()
	test1.ExpectQuery().
		WithArgs(expEmail).
		WillReturnRows(rowsPost).
		RowsWillBeClosed()
	tu := &testUUID{}
	test2.ExpectExec().
		WithArgs(expEmail, tu, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	loginStmt, err = d.Prepare(`
		SELECT 
			password, g_id, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	loginRefreshStmt, err = d.Prepare(`
			INSERT INTO user_devices (
				user_id, verifier, expires_in
			)
			VALUES
				($1, $2, $3)`)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)
	id, err := db.Login(ctx, expEmail, expPassword, false)
	assert.Nil(t, err, "unexpected error in a successful login test")
	assert.Equal(t, tu.uuid, id, "function should've returned a new session id")
	assert.Nil(t, mock.ExpectationsWereMet())

	var rowsGoogle = sqlmock.
		NewRows([]string{"password", "g_id", "activated"}).
		AddRow(nil, expPassword, true)

	mock.ExpectBegin()
	test1.ExpectQuery().
		WithArgs(expEmail).
		WillReturnRows(rowsGoogle).
		RowsWillBeClosed()
	test2.ExpectExec().
		WithArgs(expEmail, tu, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	id, err = db.Login(ctx, expEmail, expPassword, true)
	assert.Nil(t, err, "unexpected error in a successful login test")
	assert.Equal(t, tu.uuid, id, "function should've returned a new session id")
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestNotFoundLogin(t *testing.T) {
	ctx := context.Background()
	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer d.Close()

	db := DBInstance{d}

	test1 := mock.ExpectPrepare("SELECT")
	mock.ExpectBegin()
	test1.ExpectQuery().
		WithArgs(expEmail).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	loginStmt, err = d.Prepare(`
		SELECT 
			password, g_id, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	id, err := db.Login(ctx, expEmail, expPassword, false)
	assert.Empty(t, id, "unexpected output in a failed login test")
	assert.Equal(t, err, ErrAccountNotFound, "function should've returned an ErrAccountNotFound error")
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestNotActiveLogin(t *testing.T) {
	ctx := context.Background()

	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer d.Close()

	db := DBInstance{d}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(expPassword), 4)
	require.NoErrorf(t, err, "an error '%s' was not expected when creating a mock hashed password", err)

	var rows = sqlmock.
		NewRows([]string{"password", "g_id", "activated"}).
		AddRow(hashPass, nil, false)

	test1 := mock.ExpectPrepare("SELECT")
	mock.ExpectBegin()
	test1.ExpectQuery().
		WithArgs(expEmail).
		WillReturnRows(rows).
		RowsWillBeClosed()
	mock.ExpectRollback()

	loginStmt, err = d.Prepare(`
		SELECT 
			password, g_id, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	id, err := db.Login(ctx, expEmail, expPassword, false)
	assert.Empty(t, id, "unexpected output in a failed login test")
	assert.Equal(t, ErrAccountNotActive, err, "function should've returned ErrAccountNotActive error")
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestFailedLogins(t *testing.T) {
	ctx := context.Background()

	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer d.Close()

	db := DBInstance{d}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(expPassword[1:]), 4)
	require.NoErrorf(t, err, "an error '%s' was not expected when creating a mock hashed password", err)

	var rowsPost = sqlmock.
		NewRows([]string{"password", "g_id", "activated"}).
		AddRow(hashPass, nil, true)

	test1 := mock.ExpectPrepare("SELECT")
	mock.ExpectBegin()
	test1.ExpectQuery().
		WithArgs(expEmail).
		WillReturnRows(rowsPost).
		RowsWillBeClosed()
	mock.ExpectRollback()

	loginStmt, err = d.Prepare(`
		SELECT 
			password, g_id, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	id, err := db.Login(ctx, expEmail, expPassword, false)
	assert.Empty(t, id, "unexpected output in a failed login test")
	assert.Equal(t, err, ErrWrongPass, "function should've returned ErrWrongPass error")
	assert.Nil(t, mock.ExpectationsWereMet())

	var rowsGoogle = sqlmock.
		NewRows([]string{"password", "g_id", "activated"}).
		AddRow(nil, expPassword[1:], true)

	mock.ExpectBegin()
	test1.ExpectQuery().
		WithArgs(expEmail).
		WillReturnRows(rowsGoogle).
		RowsWillBeClosed()
	mock.ExpectRollback()

	id, err = db.Login(ctx, expEmail, expPassword, true)
	assert.Empty(t, id, "unexpected output in a failed login test")
	assert.Equal(t, ErrWrongId, err, "function should've returned ErrWrongId error")
	assert.Nil(t, mock.ExpectationsWereMet())
}
