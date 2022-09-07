package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserAccountInterface interface {
	Login(ctx context.Context, email string, pass string, sessionLength time.Duration) (sessionID string, err error)
	LoginGoogle(ctx context.Context, email string, gID string, sessionLength time.Duration) (sessionID string, err error)
	Register(ctx context.Context, email string, password string, name string) (activationToken string, validUntil *time.Time, err error)
	RegisterGoogle(ctx context.Context, email string, gID string, name string) (activationToken string, validUntil *time.Time, err error)
}

var loginStmt = dbStatement{
	nil, `
	SELECT 
		password, activated
	FROM 
		user_account 
	WHERE 
		email = $1`,
}
var loginGoogleStmt = dbStatement{
	nil, `
	SELECT 
		g_id, activated
	FROM 
		user_account 
	WHERE 
		email = $1`,
}
var loginRefreshStmt = dbStatement{
	nil, `
	INSERT INTO user_session (
		user_id, session_token, expires_in
	)
	VALUES
		($1, $2, $3)`,
}

var registerStmt = dbStatement{
	nil, `
	INSERT INTO user_account (
		email, password, activation_token, expires_in, name
	)
	VALUES
		($1, $2, $3, $4, $5)`,
}

var registerGoogleStmt = dbStatement{
	nil, `
	INSERT INTO user_account (
		email, g_id, activation_token, expires_in, name
	)
	VALUES
		($1, $2, $3, $4, $5)`,
}

var refreshActivationStmt = dbStatement{
	nil, `
	UPDATE user_account
	SET 
		activation_token = $1,
		expires_in = $2
	WHERE
		email = $3`,
}

func init() {
	prepareStatements = append(prepareStatements,
		&loginStmt,
		&loginGoogleStmt,
		&loginRefreshStmt,
		&registerStmt,
		&registerGoogleStmt,
		&refreshActivationStmt,
	)
}

var ErrAccountNotActive error = errors.New("account not activated")
var ErrAccountNotFound error = errors.New("account not found")
var ErrWrongID error = errors.New("google account id invalid")
var ErrWrongPass error = errors.New("account password invalid")

func (db DBInstance) Login(ctx context.Context, email string, pass string, sessionLength time.Duration) (sessionID string, err error) {
	var hash sql.NullString
	var activated bool

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	row := tx.StmtContext(ctx, loginStmt.Statement).QueryRowContext(ctx, email)
	if err := row.Scan(&hash, &activated); err != nil {
		if err == sql.ErrNoRows {
			return "", ErrAccountNotFound
		}
		return "", err
	}

	if !hash.Valid {
		return "", ErrAccountNotFound
	}

	if !activated {
		return "", ErrAccountNotActive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash.String), []byte(pass)); err != nil {
		return "", ErrWrongPass
	}

	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	sessionID = randomUUID.String()
	expiresIn := time.Now().Add(sessionLength)

	if _, err := tx.
		StmtContext(ctx, loginRefreshStmt.Statement).
		ExecContext(ctx,
			email,
			randomUUID,
			expiresIn.Format(time.RFC3339),
		); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return
}

func (db DBInstance) LoginGoogle(ctx context.Context, email string, gID string, sessionLength time.Duration) (sessionID string, err error) {
	var gid sql.NullString
	var activated bool

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	row := tx.StmtContext(ctx, loginGoogleStmt.Statement).QueryRowContext(ctx, email)
	if err := row.Scan(&gid, &activated); err != nil {
		if err == sql.ErrNoRows {
			return "", ErrAccountNotFound
		}
		return "", err
	}

	if !gid.Valid {
		return "", ErrAccountNotFound
	}

	if !activated {
		return "", ErrAccountNotActive
	}

	if gid.String != gID {
		return "", ErrWrongID
	}

	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	sessionID = randomUUID.String()
	expiresIn := time.Now().Add(sessionLength)

	if _, err := tx.
		StmtContext(ctx, loginRefreshStmt.Statement).
		ExecContext(ctx,
			email,
			randomUUID,
			expiresIn.Format(time.RFC3339),
		); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return
}

var ErrAccountExisted error = errors.New("account already existed")

func (db DBInstance) Register(ctx context.Context, email string, password string, name string) (activationToken string, validUntil *time.Time, err error) {
	row := loginStmt.Statement.QueryRowContext(ctx, email)
	if row.Err() == nil {
		var hash sql.NullString
		var activated bool
		row.Scan(&hash, &activated)
		if hash.Valid {
			return "", nil, ErrAccountExisted
		}
	} else if row.Err() != sql.ErrNoRows {
		return "", nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, err
	}

	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return "", nil, err
	}
	activationToken = randomUUID.String()

	v := time.Now().Add(time.Minute * time.Duration(2))
	validUntil = &v

	_, err = registerStmt.Statement.ExecContext(ctx, email, hash, randomUUID, validUntil, name)
	if err != nil {
		return "", nil, err
	}
	return
}

func (db DBInstance) RegisterGoogle(ctx context.Context, email string, gID string, name string) (activationToken string, validUntil *time.Time, err error) {
	row := loginGoogleStmt.Statement.QueryRowContext(ctx, email)
	if row.Err() == nil {
		var gID sql.NullString
		var activated bool
		row.Scan(&gID, &activated)
		if gID.Valid {
			return "", nil, ErrAccountExisted
		}
	} else if row.Err() != sql.ErrNoRows {
		return "", nil, err
	}

	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return "", nil, err
	}
	activationToken = randomUUID.String()

	v := time.Now().Add(time.Minute * time.Duration(2))
	validUntil = &v

	_, err = registerGoogleStmt.Statement.ExecContext(ctx, email, gID, randomUUID, validUntil, name)
	if err != nil {
		return "", nil, err
	}

	return
}

func (db DBInstance) RefreshActivation(ctx context.Context, email string) (activationToken string, validUntil *time.Time, err error) {
	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return "", nil, err
	}
	activationToken = randomUUID.String()

	v := time.Now().Add(time.Minute * time.Duration(2))
	validUntil = &v
	_, err = refreshActivationStmt.Statement.ExecContext(ctx, randomUUID, *validUntil, email)
	if err != nil {
		return "", nil, err
	}
	return
}
