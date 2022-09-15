package database

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type UserSessionInterface interface {
	GetSession(ctx context.Context, userID string, sessionToken string, currTime time.Time) (tokenFamily string, exhausted bool, expiresIn *time.Time, err error)
	AddNewSession(ctx context.Context, userID string, refreshToken string, tokenFamily string, expiresIn time.Time) error
	InvaildateSession(ctx context.Context, userID string, tokenFamily string) error
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

var ErrSessionNotFound = errors.New("")

func (db DBInstance) GetSession(ctx context.Context, userID string, sessionToken string, currTime time.Time) (tokenFamily string, exhausted bool, expiresIn *time.Time, err error) {
	if err = getRefreshStmt.Statement.QueryRowContext(ctx, userID, sessionToken, currTime).Scan(&tokenFamily, &exhausted, &expiresIn); err != nil {
		if err == sql.ErrNoRows {
			return "", false, nil, ErrSessionNotFound
		}
		return "", false, nil, err
	}
	return
}

func (db DBInstance) AddNewSession(ctx context.Context, userID string, refreshToken string, tokenFamily string, expiresIn time.Time) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.StmtContext(ctx, addRefreshStmt.Statement).Exec(userID, refreshToken, tokenFamily, expiresIn); err != nil {
		return err
	}
	if _, err := exhaustRefreshStmt.Statement.Exec(userID, refreshToken); err != nil {
		return err
	}

	return tx.Commit()
}

func (db DBInstance) InvaildateSession(ctx context.Context, userID string, tokenFamily string) error {
	_, err := invalidateTokenFamilyStmt.Statement.Exec(userID, tokenFamily)
	return err
}

func (db DBInstance) DeleteExpiredSession(ctx context.Context, currTime time.Time) (deleted int64, err error) {
	result, err := deleteExpiredRefreshStmt.Statement.ExecContext(ctx, currTime)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
