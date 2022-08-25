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
	LoginGoogle(ctx context.Context, email string, pass string, sessionLength time.Duration) (sessionID string, err error)
	Register(ctx context.Context, email string, password string, name string, viaGoogle bool) (activationToken string, validUntil *time.Time, err error)
}

var loginStmt *sql.Stmt
var loginGoogleStmt *sql.Stmt
var loginRefreshStmt *sql.Stmt
var registerStmt *sql.Stmt
var refreshActivationStmt *sql.Stmt

func init() {
	prepareStatements = append(prepareStatements,
		DBStatement{
			loginStmt, `
			SELECT 
				password, activated
			FROM 
				user_account 
			WHERE 
				email = $1`,
		},
		DBStatement{
			loginGoogleStmt, `
			SELECT 
				g_id, activated
			FROM 
				user_account 
			WHERE 
				email = $1`,
		},
		DBStatement{
			loginRefreshStmt, `
			INSERT INTO user_devices (
				user_id, verifier, expires_in
			)
			VALUES
				($1, $2, $3)`,
		},
		DBStatement{
			registerStmt, `
			INSERT INTO user_account (
				email, password, g_id, name
			)
			VALUES
				($1, $2, $3, $4)`,
		},
		DBStatement{
			refreshActivationStmt, `
			UPDATE user_account
			SET 
				activation_token = $1,
				expires_in = $2
			WHERE
				email = $3`,
		},
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

	row := tx.StmtContext(ctx, loginStmt).QueryRowContext(ctx, email)
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

	sessionID = uuid.NewString()
	expiresIn := time.Now().Add(sessionLength)

	if _, err := tx.
		StmtContext(ctx, loginRefreshStmt).
		ExecContext(ctx,
			email,
			sessionID,
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

	row := tx.StmtContext(ctx, loginGoogleStmt).QueryRowContext(ctx, email)
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

	sessionID = uuid.NewString()
	expiresIn := time.Now().Add(sessionLength)

	if _, err := tx.
		StmtContext(ctx, loginRefreshStmt).
		ExecContext(ctx,
			email,
			sessionID,
			expiresIn.Format(time.RFC3339),
		); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return
}

func (db DBInstance) Register(ctx context.Context, email string, password string, name string, viaGoogle bool) (activationToken string, validUntil *time.Time, err error) {
	if viaGoogle {
		_, err := registerStmt.ExecContext(ctx, email, nil, password, name)
		if err != nil {
			return "", nil, err
		}
	} else {
		var hash []byte
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return "", nil, err
		}
		_, err = registerStmt.ExecContext(ctx, email, hash, nil, name)
		if err != nil {
			return "", nil, err
		}
	}
	return db.RefreshActivation(ctx, email)
}

func (db DBInstance) RefreshActivation(ctx context.Context, email string) (activationToken string, validUntil *time.Time, err error) {
	activationToken = uuid.NewString()
	*validUntil = time.Now().Add(time.Minute * time.Duration(2))
	_, err = refreshActivationStmt.ExecContext(ctx, activationToken, validUntil, email)
	return
}
