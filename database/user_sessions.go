package database

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type UserSessionInterface interface {
	CheckSession(ctx context.Context, userID string, sessionToken string, currTime time.Time) (isValidSession bool, err error)
}

var isValidSessionQuery = dbStatement{
	nil, `
	SELECT
		EXISTS(
			SELECT
			1
			FROM
				user_session
			WHERE
				user_id = $1
				AND session_token = $2
				AND expires_in >= $3
		) AS session_valid;`,
}

func init() {
	prepareStatements = append(prepareStatements,
		&isValidSessionQuery,
	)
}

var ErrSessionInvalid = errors.New("")

func (db DBInstance) CheckSession(ctx context.Context, userID string, sessionToken string, currTime time.Time) (isSessionValid bool, err error) {
	row := isValidSessionQuery.Statement.QueryRowContext(ctx, userID, sessionToken, currTime)
	if err = row.Scan(&isSessionValid); err != nil {
		if err == sql.ErrNoRows {
			return false, ErrSessionInvalid
		}
		return false, err
	}

	if !isSessionValid {
		err = ErrSessionInvalid
	}
	return
}
