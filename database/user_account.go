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
	Login(ctx context.Context, email string, pass string, sessionLength time.Duration) (refreshID string, err error)
	LoginGoogle(ctx context.Context, email string, gID string, sessionLength time.Duration) (refreshID string, err error)
	Register(ctx context.Context, email string, password string, name string, activationDuration time.Duration) (activationToken string, validUntil *time.Time, err error)
	RegisterGoogle(ctx context.Context, email string, gID string, name string, activationDuration time.Duration) (activationToken string, validUntil *time.Time, err error)
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

var registerSearchStmt = dbStatement{
	nil, `
	SELECT 
		password, activated
	FROM 
		user_account 
	WHERE 
		email = $1
	FOR UPDATE`,
}
var registerSearchGoogleStmt = dbStatement{
	nil, `
	SELECT 
		g_id, activated
	FROM 
		user_account 
	WHERE 
		email = $1
	FOR UPDATE`,
}
var registerStmt = dbStatement{
	nil, `
	INSERT INTO user_account (
		email, password, activation_token, expires_in, name
	)
	VALUES
		($1, $2, $3, $4, $5)`,
}
var registerAddStmt = dbStatement{
	nil, `
	UPDATE user_account
	SET
		password = $2,
		activation_token = $3,
		expires_in = $4
	WHERE
		email = $1`,
}
var registerGoogleStmt = dbStatement{
	nil, `
	INSERT INTO user_account (
		email, g_id, activation_token, expires_in, name
	)
	VALUES
		($1, $2, $3, $4, $5)`,
}
var registerAddGoogleStmt = dbStatement{
	nil, `
	UPDATE user_account
	SET
		g_id = $2,
		activation_token = $3,
		expires_in = $4
	WHERE
		email = $1`,
}

var refreshActivationStmt = dbStatement{
	nil, `
	UPDATE user_account
	SET 
		activated = $1,
		activation_token = $2,
		expires_in = $3
	WHERE
		email = $4`,
}

func init() {
	prepareStatements = append(prepareStatements,
		&loginStmt,
		&loginGoogleStmt,
		&registerSearchStmt,
		&registerSearchGoogleStmt,
		&registerStmt,
		&registerAddStmt,
		&registerGoogleStmt,
		&registerAddGoogleStmt,
		&refreshActivationStmt,
	)
}

var ErrAccountNotActive error = errors.New("account not activated")
var ErrAccountNotFound error = errors.New("account not found")
var ErrWrongID error = errors.New("google account id invalid")
var ErrWrongPass error = errors.New("account password invalid")

func (db DBInstance) Login(ctx context.Context, email string, pass string, sessionLength time.Duration) (refreshID string, err error) {
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
	refreshID = randomUUID.String()
	expiresIn := time.Now().Add(sessionLength)

	if _, err := tx.
		StmtContext(ctx, addRefreshStmt.Statement).
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

func (db DBInstance) LoginGoogle(ctx context.Context, email string, gID string, sessionLength time.Duration) (refreshID string, err error) {
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
	refreshID = randomUUID.String()
	expiresIn := time.Now().Add(sessionLength)

	if _, err := tx.
		StmtContext(ctx, addRefreshStmt.Statement).
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

func (db DBInstance) Register(ctx context.Context, email string, password string, name string, activationDuration time.Duration) (activationToken string, validUntil *time.Time, err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", nil, err
	}
	defer tx.Rollback()

	row := tx.StmtContext(ctx, registerSearchStmt.Statement).QueryRowContext(ctx, email)
	var nullHash sql.NullString
	var activated bool

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, err
	}

	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return "", nil, err
	}
	activationToken = randomUUID.String()

	v := time.Now().Add(activationDuration)
	validUntil = &v

	err = row.Scan(&nullHash, &activated)
	if err == nil {
		if nullHash.Valid || !activated {
			return "", nil, ErrAccountExisted
		}
		_, err = tx.StmtContext(ctx, registerAddStmt.Statement).ExecContext(ctx, email, hash, randomUUID, validUntil)
		if err != nil {
			return "", nil, err
		}
	} else if err != sql.ErrNoRows {
		return "", nil, err
	} else {
		_, err = tx.StmtContext(ctx, registerStmt.Statement).ExecContext(ctx, email, hash, randomUUID, validUntil, name)
		if err != nil {
			return "", nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", nil, err
	}
	return
}

func (db DBInstance) RegisterGoogle(ctx context.Context, email string, gID string, name string, activationDuration time.Duration) (activationToken string, validUntil *time.Time, err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", nil, err
	}
	defer tx.Rollback()

	row := tx.StmtContext(ctx, registerSearchGoogleStmt.Statement).QueryRowContext(ctx, email)
	var nullGID sql.NullString
	var activated bool

	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return "", nil, err
	}
	activationToken = randomUUID.String()

	v := time.Now().Add(activationDuration)
	validUntil = &v

	err = row.Scan(&nullGID, &activated)
	if err == nil {
		if nullGID.Valid || !activated {
			return "", nil, ErrAccountExisted
		}
		_, err = tx.StmtContext(ctx, registerAddGoogleStmt.Statement).ExecContext(ctx, email, gID, randomUUID, validUntil)
		if err != nil {
			return "", nil, err
		}
	} else if err != sql.ErrNoRows {
		return "", nil, err
	} else {
		_, err = tx.StmtContext(ctx, registerGoogleStmt.Statement).ExecContext(ctx, email, gID, randomUUID, validUntil, name)
		if err != nil {
			return "", nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", nil, err
	}
	return
}

func (db DBInstance) RefreshActivation(ctx context.Context, email string, duration time.Duration) (activationToken string, validUntil *time.Time, err error) {
	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return "", nil, err
	}
	activationToken = randomUUID.String()

	v := time.Now().Add(duration)
	validUntil = &v
	res, err := refreshActivationStmt.Statement.ExecContext(ctx, false, randomUUID, *validUntil, email)
	if err != nil {
		return "", nil, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return "", nil, err
	}
	if rowsAffected == 0 {
		return "", nil, ErrAccountNotFound
	}
	return
}
