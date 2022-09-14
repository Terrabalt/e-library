package database

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type UserSessionInterface interface {
	CheckSession(ctx context.Context, userID string, sessionToken string, currTime time.Time) (newRefresh string, err error)
}

var getRefreshStmt = dbStatement{
	nil, `
	SELECT
		token_family, exhausted, expires_in
	FROM
		user_session
	WHERE
		user_id = $1
		AND refresh_token = $2;`,
}
var exhaustRefreshStmt = dbStatement{
	nil, `
	UPDATE user_session
	SET
		exhausted = 'false'
	WHERE
		user_id = $1
		AND refresh_token = $2`,
}
var addRefreshStmt = dbStatement{
	nil, `
	INSERT INTO user_session (
		user_id, refresh_token, expires_in
	)
	VALUES
		($1, $2, $3)`,
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
		&deleteExpiredRefreshStmt,
	)
}

var ErrSessionInvalid = errors.New("")

func (db DBInstance) CheckSession(ctx context.Context, userID string, sessionToken string, currTime time.Time) (newRefresh string, err error) {
	tx, err := db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	row := tx.Stmt(getRefreshStmt.Statement).QueryRowContext(ctx, userID, sessionToken, currTime)
	var s string
	var exhausted bool
	var expiresIn time.Time

	if err = row.Scan(&s, &exhausted, &expiresIn); err != nil {
		if err == sql.ErrNoRows {
			return "", ErrSessionInvalid
		}
		return "", err
	}

	if exhausted || expiresIn.Before(time.Now()) {
		err = ErrSessionInvalid
	}

	tx.Commit()
	return
}

func (db DBInstance) DeleteExpiredSession(ctx context.Context, currTime time.Time) (deleted int64, err error) {
	result, err := deleteExpiredRefreshStmt.Statement.ExecContext(ctx, currTime)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
