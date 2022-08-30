package database

import (
	"context"
	"database/sql"
)

type dbStatement struct {
	Statement *sql.Stmt
	Query     string
}

func (stmt *dbStatement) Prepare(ctx context.Context, db *sql.DB) (err error) {
	(*stmt).Statement, err = db.PrepareContext(ctx, stmt.Query)
	return
}

func (stmt *dbStatement) Close() (err error) {
	return (*stmt).Statement.Close()
}
