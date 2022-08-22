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
	Login(ctx context.Context, email string, pass string) (id string, err error)
}

var loginStmt *sql.Stmt
var loginRefreshStmt *sql.Stmt

func init() {
	prepareStatements = append(prepareStatements,
		DBStatement{
			&loginStmt, `
			SELECT 
				password, activated
			FROM 
				user_account 
			WHERE 
				email = $1`,
		},
		DBStatement{
			&loginStmt, `
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

/// Logs the user in, and returns a new identifier with it
func (db DBInstance) Login(ctx context.Context, email string, pass string) (id string, err error) {
	var hash string
	var activated bool

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	cursor := tx.StmtContext(ctx, loginStmt).QueryRowContext(ctx, email)
	if err := cursor.Scan(&hash, &activated); err != nil {
		return "", err
	}

	if !activated {
		return "", ErrAccountNotActive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass)); err != nil {
		return "", err
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
