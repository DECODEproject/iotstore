package postgres

import (
	kitlog "github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DB struct {
	connStr string
	logger  kitlog.Logger
	db      *sqlx.DB
}

func NewDB(connStr string, logger kitlog.Logger) *DB {
	logger = kitlog.With(logger, "module", "postgres")

	return &DB{
		connStr: connStr,
		logger:  logger,
	}
}

func Open(connStr string) (*sqlx.DB, error) {
	return sqlx.Open("postgres", connStr)
}

func (d *DB) Start() error {
	d.logger.Log("msg", "starting postgres")

	db, err := Open(d.connStr)
	if err != nil {
		return err
	}

	d.db = db

	err = MigrateUp(d.db.DB, d.logger)
	if err != nil {
		return err
	}

	return nil
}

func (d *DB) Stop() error {
	d.logger.Log("msg", "stopping postgres")

	return d.db.Close()
}
