package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type UserSessionInterface interface {
	CheckSession(ctx context.Context, userID string, sessionToken string, currTime time.Time, sessionLength time.Duration) (newRefresh string, err error)
}

var getRefreshStmt = dbStatement{
	nil, `
	SELECT
		token_family, exhausted, expires_in
	FROM
		user_session
	WHERE
		user_id = $1
		AND refresh_token = $2
	FOR UPDATE;`,
}
var exhaustRefreshStmt = dbStatement{
	nil, `
	UPDATE user_session
	SET
		exhausted = 'true'
	WHERE
		user_id = $1
		AND refresh_token = $2;`,
}
var addRefreshStmt = dbStatement{
	nil, `
	INSERT INTO user_session (
		user_id, refresh_token, token_family, expires_in
	)
	VALUES
		($1, $2, $3, $4);`,
}
var invalidateTokenFamilyStmt = dbStatement{
	nil, `
	DELETE FROM
		user_session
	WHERE
		user_id = $1
		AND token_family = $2`,
}

var deleteExpiredRefreshStmt = dbStatement{
	nil, `
	DELETE FROM
		user_session
	WHERE
		expires_in <= $1;`,
}

func init() {
	prepareStatements = append(prepareStatements,
		&getRefreshStmt,
		&exhaustRefreshStmt,
		&addRefreshStmt,
		&invalidateTokenFamilyStmt,
		&deleteExpiredRefreshStmt,
	)
}

var ErrSessionInvalid = errors.New("")

func (db DBInstance) CheckSession(ctx context.Context, userID string, sessionToken string, currTime time.Time, sessionLength time.Duration) (newRefreshToken string, err error) {
	tx, err := db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	row := tx.StmtContext(ctx, getRefreshStmt.Statement).QueryRowContext(ctx, userID, sessionToken, currTime)
	var tokenFamily string
	var exhausted bool
	var expiresIn time.Time

	if err = row.Scan(&tokenFamily, &exhausted, &expiresIn); err != nil {
		if err == sql.ErrNoRows {
			return "", ErrSessionInvalid
		}
		return "", err
	}

	if expiresIn.Before(time.Now()) {
		return "", ErrSessionInvalid
	}

	if exhausted {
		_, err := tx.StmtContext(ctx, invalidateTokenFamilyStmt.Statement).ExecContext(ctx, userID, tokenFamily)
		if err != nil {
			return "", err
		}
		if err := tx.Commit(); err != nil {
			return "", err
		}
		return "", ErrSessionInvalid
	}

	newRefresh, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	expiresIn = time.Now().Add(sessionLength)
	if _, err := tx.StmtContext(ctx, addRefreshStmt.Statement).
		Exec(userID, newRefresh, tokenFamily, expiresIn); err != nil {
		return "", err
	}

	if _, err := tx.StmtContext(ctx, exhaustRefreshStmt.Statement).
		Exec(userID, sessionToken); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return newRefresh.String(), nil
}

func (db DBInstance) DeleteExpiredSession(ctx context.Context, currTime time.Time) (deleted int64, err error) {
	result, err := deleteExpiredRefreshStmt.Statement.ExecContext(ctx, currTime)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
