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

type testString struct {
	str string
}

func (tu *testString) Match(value driver.Value) bool {
	switch v := value.(type) {
	case []byte:
		tu.str = string(v)
		return true
	case string:
		tu.str = v
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

	loginStmt, err = d.Prepare(`
		SELECT 
			password, activated
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

	loginStmt, err = d.Prepare(`
		SELECT 
			password, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
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

	loginStmt, err = d.Prepare(`
		SELECT 
			password, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
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

	loginStmt, err = d.Prepare(`
		SELECT 
			password, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
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

	loginGoogleStmt, err = d.Prepare(`
		SELECT 
			g_id, activated
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

	loginGoogleStmt, err = d.Prepare(`
		SELECT 
			g_id, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
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

	loginGoogleStmt, err = d.Prepare(`
		SELECT 
			g_id, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
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

	loginGoogleStmt, err = d.Prepare(`
		SELECT 
			g_id, activated
		FROM 
			user_account 
		WHERE 
			email = $1`)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	id, err := db.LoginGoogle(ctx, expEmail, expGID, time.Duration(48)*time.Hour)
	assert.Empty(t, id, "unexpected output in a failed login test")
	assert.Equal(t, ErrWrongID, err, "function should've returned ErrWrongPass error")
	assert.Nil(t, mock.ExpectationsWereMet())
}

const expName = "Joko"

func TestSuccessfulRegister(t *testing.T) {
	ctx := context.Background()

	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer d.Close()

	db := DBInstance{d}

	th := &testString{}
	test1 := mock.ExpectPrepare("INSERT")
	test2 := mock.ExpectPrepare("UPDATE")
	test1.ExpectExec().
		WithArgs(expEmail, th, expName).
		WillReturnResult(sqlmock.NewResult(0, 1))
	tu := &testUUID{}
	test2.ExpectExec().
		WithArgs(tu, sqlmock.AnyArg(), expEmail).
		WillReturnResult(sqlmock.NewResult(0, 1))

	registerStmt, err = d.Prepare(`
		INSERT INTO user_account (
			email, password, name
		)
		VALUES
			($1, $2, $3)`)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	refreshActivationStmt, err = d.Prepare(`
		UPDATE user_account
		SET 
			activation_token = $1,
			expires_in = $2
		WHERE
			email = $3`)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)
	actToken, _, err := db.Register(ctx, expEmail, expPassword, expName)
	assert.Nil(t, err, "unexpected error in a successful login test")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(th.str), []byte(expPassword)), "function should've returned a correctly-hashed password")
	assert.Equal(t, tu.uuid, actToken, "function should've returned a new session id")
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestSuccessfulRegisterGoogle(t *testing.T) {
	ctx := context.Background()

	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer d.Close()

	db := DBInstance{d}

	th := &testString{}
	test1 := mock.ExpectPrepare("INSERT")
	test2 := mock.ExpectPrepare("UPDATE")
	test1.ExpectExec().
		WithArgs(expEmail, th, expName).
		WillReturnResult(sqlmock.NewResult(0, 1))
	tu := &testUUID{}
	test2.ExpectExec().
		WithArgs(tu, sqlmock.AnyArg(), expEmail).
		WillReturnResult(sqlmock.NewResult(0, 1))

	registerGoogleStmt, err = d.Prepare(`
		INSERT INTO user_account (
			email, g_id, name
		)
		VALUES
			($1, $2, $3)`)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	refreshActivationStmt, err = d.Prepare(`
		UPDATE user_account
		SET 
			activation_token = $1,
			expires_in = $2
		WHERE
			email = $3`)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)
	actToken, _, err := db.RegisterGoogle(ctx, expEmail, expGID, expName)
	assert.Nil(t, err, "unexpected error in a successful login test")
	assert.Equal(t, expGID, th.str, "function should've returned a correct google account id")
	assert.Equal(t, tu.uuid, actToken, "function should've returned a new session id")
	assert.Nil(t, mock.ExpectationsWereMet())
}
