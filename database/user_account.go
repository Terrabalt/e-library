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
	Login(ctx context.Context, email string, pass string, viaGoogle bool) (id string, err error)
}

var loginStmt *sql.Stmt
var loginRefreshStmt *sql.Stmt

func init() {
	prepareStatements = append(prepareStatements,
		DBStatement{
			&loginStmt, `
			SELECT 
				password, g_id, activated
			FROM 
				user_account 
			WHERE 
				email = $1`,
		},
		DBStatement{
			&loginRefreshStmt, `
			INSERT INTO user_devices (
				user_id, verifier, expires_in
			)
			VALUES
				($1, $2, $3)`,
		},
	)
}

var ErrAccountNotActive error = errors.New("account not activated")
var ErrAccountNotFound error = errors.New("account not found")
var ErrWrongId error = errors.New("google account id invalid")
var ErrWrongPass error = errors.New("account password invalid")

/// Logs the user in, and returns a new identifier with it
func (db DBInstance) Login(ctx context.Context, email string, pass string, viaGoogle bool) (id string, err error) {
	var hash, gid sql.NullString
	var activated bool

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	cursor := tx.StmtContext(ctx, loginStmt).QueryRowContext(ctx, email)
	if err := cursor.Scan(&hash, &gid, &activated); err != nil {
		return "", ErrAccountNotFound
	}

	if !activated {
		return "", ErrAccountNotActive
	}

	if viaGoogle {
		if !gid.Valid {
			return "", ErrAccountNotActive
		}
		if gid.String != pass {
			return "", ErrWrongId
		}
	} else {
		if !hash.Valid {
			return "", ErrAccountNotActive
		}
		if err := bcrypt.CompareHashAndPassword([]byte(hash.String), []byte(pass)); err != nil {
			log.Debug().Caller().Err(err).Msg("Error when comparing password to hash")
			return "", ErrWrongPass
		}
	}

	verifier := uuid.NewString()
	expiresIn := time.Now().Add(time.Duration(48) * time.Hour)

	if _, err := tx.
		StmtContext(ctx, loginRefreshStmt).
		ExecContext(ctx,
			email,
			verifier,
			expiresIn.Format(time.RFC3339),
		); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return verifier, nil
}
