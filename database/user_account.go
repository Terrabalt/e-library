package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type UserAccountInterface interface {
	Login(ctx context.Context, email string, pass string, sessionLength time.Duration) (sessionID string, err error)
	LoginGoogle(ctx context.Context, email string, gID string, sessionLength time.Duration) (sessionID string, err error)
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
	INSERT INTO user_devices (
		user_id, verifier, expires_in
	)
	VALUES
		($1, $2, $3)`,
}

func init() {
	prepareStatements = append(prepareStatements,
		&loginStmt,
		&loginGoogleStmt,
		&loginRefreshStmt,
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
		log.Debug().Err(err).Msg("Error when comparing password to hash")
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
