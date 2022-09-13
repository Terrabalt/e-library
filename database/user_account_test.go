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

	err = loginStmt.Prepare(ctx, d)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	err = loginRefreshStmt.Prepare(ctx, d)
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

	err = loginStmt.Prepare(ctx, d)
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

	err = loginStmt.Prepare(ctx, d)
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

	err = loginStmt.Prepare(ctx, d)
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

	err = loginGoogleStmt.Prepare(ctx, d)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	err = loginRefreshStmt.Prepare(ctx, d)
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

	err = loginGoogleStmt.Prepare(ctx, d)
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

	err = loginGoogleStmt.Prepare(ctx, d)
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

	err = loginGoogleStmt.Prepare(ctx, d)
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

	expDur := time.Minute * time.Duration(10)

	db := DBInstance{d}

	th := &testString{}
	test1 := mock.ExpectPrepare("SELECT")
	test2 := mock.ExpectPrepare("INSERT")
	mock.ExpectBegin()
	test1.ExpectQuery().
		WithArgs(expEmail).
		WillReturnError(sql.ErrNoRows)
	tu := &testUUID{}
	test2.ExpectExec().
		WithArgs(expEmail, th, tu, sqlmock.AnyArg(), expName).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = registersearchStmt.Prepare(ctx, d)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	err = registerStmt.Prepare(ctx, d)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	actToken, _, err := db.Register(ctx, expEmail, expPassword, expName, expDur)
	if assert.Nil(t, err, "unexpected error in a successful register test") {
		assert.Equal(t, tu.uuid, actToken, "function should've returned a new session id")
	}
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(th.str), []byte(expPassword)), "function should've returned a correctly-hashed password")
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestSuccessfulAddRegister(t *testing.T) {
	ctx := context.Background()

	var rowsPost = sqlmock.
		NewRows([]string{"password", "activated"}).
		AddRow(nil, true)

	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer d.Close()

	expDur := time.Minute * time.Duration(10)

	db := DBInstance{d}

	th := &testString{}
	test1 := mock.ExpectPrepare("SELECT")
	test2 := mock.ExpectPrepare("UPDATE")
	mock.ExpectBegin()
	test1.ExpectQuery().
		WithArgs(expEmail).
		WillReturnRows(rowsPost)
	tu := &testUUID{}
	test2.ExpectExec().
		WithArgs(expEmail, th, tu, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = registersearchStmt.Prepare(ctx, d)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	err = registerAddStmt.Prepare(ctx, d)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	actToken, _, err := db.Register(ctx, expEmail, expPassword, expName, expDur)
	if assert.Nil(t, err, "unexpected error in a successful register test") {
		assert.Equal(t, tu.uuid, actToken, "function should've returned a new session id")
	}
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(th.str), []byte(expPassword)), "function should've returned a correctly-hashed password")
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestSuccessfulRegisterGoogle(t *testing.T) {
	ctx := context.Background()

	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer d.Close()
	expDur := time.Minute * time.Duration(10)

	db := DBInstance{d}

	th := &testString{}
	test1 := mock.ExpectPrepare("SELECT")
	test2 := mock.ExpectPrepare("INSERT")
	mock.ExpectBegin()
	test1.ExpectQuery().
		WithArgs(expEmail).
		WillReturnError(sql.ErrNoRows)
	tu := &testUUID{}
	test2.ExpectExec().
		WithArgs(expEmail, th, tu, sqlmock.AnyArg(), expName).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = registersearchGoogleStmt.Prepare(ctx, d)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	err = registerGoogleStmt.Prepare(ctx, d)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	actToken, _, err := db.RegisterGoogle(ctx, expEmail, expGID, expName, expDur)
	if assert.Nil(t, err, "unexpected error in a successful register-google test") {
		assert.Equal(t, tu.uuid, actToken, "function should've returned a new session id")
	}
	assert.Equal(t, expGID, th.str, "function should've returned a correct google account id")
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestSuccessfulAddRegisterGoogle(t *testing.T) {
	ctx := context.Background()

	var rowsGoogle = sqlmock.
		NewRows([]string{"g_id", "activated"}).
		AddRow(nil, true)

	d, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer d.Close()
	expDur := time.Minute * time.Duration(10)

	db := DBInstance{d}

	th := &testString{}
	test1 := mock.ExpectPrepare("SELECT")
	test2 := mock.ExpectPrepare("UPDATE")
	mock.ExpectBegin()
	test1.ExpectQuery().
		WithArgs(expEmail).
		WillReturnRows(rowsGoogle)
	tu := &testUUID{}
	test2.ExpectExec().
		WithArgs(expEmail, th, tu, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = registersearchGoogleStmt.Prepare(ctx, d)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	err = registerAddGoogleStmt.Prepare(ctx, d)
	require.NoErrorf(t, err, "an error '%s' was not expected when preparing a stub database connection", err)

	actToken, _, err := db.RegisterGoogle(ctx, expEmail, expGID, expName, expDur)
	if assert.Nil(t, err, "unexpected error in a successful register-google test") {
		assert.Equal(t, tu.uuid, actToken, "function should've returned a new session id")
	}
	assert.Equal(t, expGID, th.str, "function should've returned a correct google account id")
	assert.Nil(t, mock.ExpectationsWereMet())
}
