package database

import (
	"context"
	"database/sql"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"
)

type DB interface {
	UserAccountInterface
	UserSessionInterface
	InitDB(ctx context.Context) error
	CloseDB()
}

type DBInstance struct {
	*sql.DB
}

var prepareStatements []*dbStatement

func StartDB(dbInfo string) (DB, error) {
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return nil, err
	}

	loggerAdapter := zerologadapter.New(log.Logger)
	db = sqldblogger.OpenDriver(dbInfo, db.Driver(), loggerAdapter, sqldblogger.WithSQLQueryAsMessage(true))
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Info().Str("dsn", dbInfo).Msg("Database connected")
	return DBInstance{db}, nil
}

func (db DBInstance) InitDB(ctx context.Context) error {
	files, err := ioutil.ReadDir(filepath.Join(".", "database", "sql_versions"))
	if err != nil {
		return err
	}

	var latestVer = 0
	var latestSQLFile fs.FileInfo
	for _, file := range files {
		ver := strings.SplitN(file.Name(), "_", 2)
		currVer, err := strconv.Atoi(ver[0])

		if err != nil {
			log.Printf("error getting version on file \"%s\": %v.\n", file.Name(), err)
			return err
		} else if latestVer <= currVer {
			latestVer = currVer
			latestSQLFile = file
		}

	}

	c, err := ioutil.ReadFile(filepath.Join(".", "database", "sql_versions", latestSQLFile.Name()))
	if err != nil {
		log.Error().Err(err).Str("file", latestSQLFile.Name()).Msg("error reading sql file")
		return err
	}
	sql := string(c)
	if _, err := db.ExecContext(ctx, sql); err != nil {
		log.Error().Err(err).Str("sql", sql).Msg("error running initializing sql file")
		return err
	}

	for _, stmt := range prepareStatements {
		if err := stmt.Prepare(ctx, db.DB); err != nil {
			log.Error().Err(err).Str("query", stmt.Query).Msg("error preparing statements")
			return err
		}
	}
	return nil
}

func (db DBInstance) CloseDB() {
	for _, stmt := range prepareStatements {
		if err := stmt.Close(); err != nil {
			log.Error().Err(err).Msg("error closing db")
		}
	}
	db.Close()
}
