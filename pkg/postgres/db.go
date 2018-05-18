package postgres

import (
	"database/sql"

	kitlog "github.com/go-kit/kit/log"
	_ "github.com/lib/pq"
)

type DB struct {
	connStr string
	logger  kitlog.Logger
	db      *sql.DB
}

func NewDB(connStr string, logger kitlog.Logger) *DB {
	logger = kitlog.With(logger, "module", "postgres")

	return &DB{
		connStr: connStr,
		logger:  logger,
	}
}

func (d *DB) Start() error {
	d.logger.Log("msg", "starting postgres")

	db, err := sql.Open("postgres", d.connStr)
	if err != nil {
		return err
	}

	err = MigrateUp(d.db, d.logger)
	if err != nil {
		return err
	}

	d.db = db

	return nil
}

func (d *DB) Stop() error {
	d.logger.Log("msg", "stopping postgres")

	return d.db.Close()
}
