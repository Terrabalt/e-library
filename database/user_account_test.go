package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

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
const expGID = "abcdefgh-ijkl"

func TestSuccessfulLogin(t *testing.T) {
	ctx := context.Background()

	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer d.Close()

	db := DBInstance{d}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(expPassword), 4)
	require.NoErrorf(t, err, "an error '%s' was not expected when creating a mock hashed password", err)

	var rowsPost = sqlmock.
		NewRows([]string{"password", "activated"}).
		AddRow(hashPass, true)

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

	stmt, err := d.Prepare(`
		SELECT 
			password, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
	loginStmt = *stmt
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	stmt, err = d.Prepare(`
			INSERT INTO user_devices (
				user_id, verifier, expires_in
			)
			VALUES
				($1, $2, $3)`)
	loginRefreshStmt = *stmt
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)
	id, err := db.Login(ctx, expEmail, expPassword, time.Duration(48)*time.Hour)
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

	stmt, err := d.Prepare(`
		SELECT 
			password, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
	loginStmt = *stmt
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	id, err := db.Login(ctx, expEmail, expPassword, time.Duration(48)*time.Hour)
	assert.Empty(t, id, "unexpected output in a failed login test")
	assert.Equal(t, ErrAccountNotFound, err, "function should've returned an ErrAccountNotFound error")
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
		NewRows([]string{"password", "activated"}).
		AddRow(hashPass, false)

	test1 := mock.ExpectPrepare("SELECT")
	mock.ExpectBegin()
	test1.ExpectQuery().
		WithArgs(expEmail).
		WillReturnRows(rows).
		RowsWillBeClosed()
	mock.ExpectRollback()

	stmt, err := d.Prepare(`
		SELECT 
			password, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
	loginStmt = *stmt
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	id, err := db.Login(ctx, expEmail, expPassword, time.Duration(48)*time.Hour)
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
		NewRows([]string{"password", "activated"}).
		AddRow(hashPass, true)

	test1 := mock.ExpectPrepare("SELECT")
	mock.ExpectBegin()
	test1.ExpectQuery().
		WithArgs(expEmail).
		WillReturnRows(rowsPost).
		RowsWillBeClosed()
	mock.ExpectRollback()

	stmt, err := d.Prepare(`
		SELECT 
			password, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
	loginStmt = *stmt
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	id, err := db.Login(ctx, expEmail, expPassword, time.Duration(48)*time.Hour)
	assert.Empty(t, id, "unexpected output in a failed login test")
	assert.Equal(t, ErrWrongPass, err, "function should've returned ErrWrongPass error")
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestSuccessfulLoginGoogle(t *testing.T) {
	ctx := context.Background()

	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer d.Close()

	db := DBInstance{d}

	var rowsPost = sqlmock.
		NewRows([]string{"g_id", "activated"}).
		AddRow(expGID, true)

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

	stmt, err := d.Prepare(`
		SELECT 
			g_id, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
	loginGoogleStmt = *stmt
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	stmt, err = d.Prepare(`
			INSERT INTO user_devices (
				user_id, verifier, expires_in
			)
			VALUES
				($1, $2, $3)`)
	loginRefreshStmt = *stmt
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	id, err := db.LoginGoogle(ctx, expEmail, expGID, time.Duration(48)*time.Hour)
	assert.Nil(t, err, "unexpected error in a successful login test")
	assert.Equal(t, tu.uuid, id, "function should've returned a new session id")
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestNotFoundLoginGoogle(t *testing.T) {
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

	stmt, err := d.Prepare(`
		SELECT 
			g_id, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
	loginGoogleStmt = *stmt
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	id, err := db.LoginGoogle(ctx, expEmail, expGID, time.Duration(48)*time.Hour)
	assert.Empty(t, id, "unexpected output in a failed login test")
	assert.Equal(t, ErrAccountNotFound, err, "function should've returned an ErrAccountNotFound error")
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestNotActiveLoginGoogle(t *testing.T) {
	ctx := context.Background()

	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer d.Close()

	db := DBInstance{d}

	var rows = sqlmock.
		NewRows([]string{"g_id", "activated"}).
		AddRow(expGID, false)

	test1 := mock.ExpectPrepare("SELECT")
	mock.ExpectBegin()
	test1.ExpectQuery().
		WithArgs(expEmail).
		WillReturnRows(rows).
		RowsWillBeClosed()
	mock.ExpectRollback()

	stmt, err := d.Prepare(`
		SELECT 
			g_id, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
	loginGoogleStmt = *stmt
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	id, err := db.LoginGoogle(ctx, expEmail, expGID, time.Duration(48)*time.Hour)
	assert.Empty(t, id, "unexpected output in a failed login test")
	assert.Equal(t, ErrAccountNotActive, err, "function should've returned ErrAccountNotActive error")
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestFailedLoginsGoogle(t *testing.T) {
	ctx := context.Background()

	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer d.Close()

	db := DBInstance{d}

	var rowsPost = sqlmock.
		NewRows([]string{"g_id", "activated"}).
		AddRow(expGID[1:], true)

	test1 := mock.ExpectPrepare("SELECT")
	mock.ExpectBegin()
	test1.ExpectQuery().
		WithArgs(expEmail).
		WillReturnRows(rowsPost).
		RowsWillBeClosed()
	mock.ExpectRollback()

	stmt, err := d.Prepare(`
		SELECT 
			g_id, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
	loginGoogleStmt = *stmt
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	id, err := db.LoginGoogle(ctx, expEmail, expGID, time.Duration(48)*time.Hour)
	assert.Empty(t, id, "unexpected output in a failed login test")
	assert.Equal(t, ErrWrongID, err, "function should've returned ErrWrongPass error")
	assert.Nil(t, mock.ExpectationsWereMet())
}
