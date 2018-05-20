package postgres

import (
	kitlog "github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

// DB is our type that wraps a sqlx.DB instance.
type DB struct {
	connStr string
	logger  kitlog.Logger
	db      *sqlx.DB
}

// NewDB takes as input a connection string, and a logger and returns an
// instantiated DB instance. The connection pool is not created at this point,
// we only create that when Start is called on the DB instance.
func NewDB(connStr string, logger kitlog.Logger) *DB {
	logger = kitlog.With(logger, "module", "postgres")

	return &DB{
		connStr: connStr,
		logger:  logger,
	}
}

// Open takes as input a connection string for a DB, and returns either a
// sqlx.DB instance or an error.
func Open(connStr string) (*sqlx.DB, error) {
	return sqlx.Open("postgres", connStr)
}

// Start is where we create the DB connection.
func (d *DB) Start() error {
	d.logger.Log("msg", "starting postgres")

	db, err := Open(d.connStr)
	if err != nil {
		return errors.Wrap(err, "opening db connection failed")
	}

	d.db = db

	err = MigrateUp(d.db.DB, d.logger)
	if err != nil {
		return errors.Wrap(err, "running up migrations failed")
	}

	return nil
}

// Stop closes the DB connection
func (d *DB) Stop() error {
	d.logger.Log("msg", "stopping postgres")

	return d.db.Close()
}
