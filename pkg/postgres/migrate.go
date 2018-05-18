package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
)

func MigrateUp(db *sql.DB, logger kitlog.Logger) error {
	logger.Log("msg", "migrating DB up")

	m, err := getMigrator(db, logger)
	if err != nil {
		return err
	}

	return m.Up()
}

func MigrateDown(db *sql.DB, logger kitlog.Logger) error {
	logger.Log("msg", "migrating DB down")

	m, err := getMigrator(db, logger)
	if err != nil {
		return err
	}

	return m.Down()
}

// NewMigration creates a new pair of files into which an SQL migration should
// be written. All this is doing is ensuring files created are correctly named.
func NewMigration(dirName, migrationName string, logger kitlog.Logger) error {
	if migrationName == "" {
		return errors.New("Must specify a name when creating a migration")
	}

	migrationID := time.Now().Format("20060102150405") + "_" + migrationName
	upFileName := fmt.Sprintf("%s_up.sql", migrationID)
	downFileName := fmt.Sprintf("%s_down.sql", migrationID)

	logger.Log("upfile", upFileName, "downfile", downFileName, "directory", dirName, "msg", "creating migration files")

	err := os.MkdirAll(dirName, 0755)
	if err != nil {
		return err
	}

	upFile, err := os.Create(filepath.Join(dirName, upFileName))
	if err != nil {
		return err
	}
	defer upFile.Close()

	downFile, err := os.Create(filepath.Join(dirName, downFileName))
	if err != nil {
		return err
	}
	defer downFile.Close()

	return nil
}

func getMigrator(db *sql.DB, logger kitlog.Logger) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file:///pkg/migrations",
		"postgres",
		driver,
	)

	if err != nil {
		return nil, err
	}

	la := newLogAdapter(logger, true)

	m.Log = la

	return m, nil
}

func newLogAdapter(logger kitlog.Logger, verbose bool) migrate.Logger {
	return &logAdapter{logger: logger, verbose: verbose}
}

// logAdapter is a simple type we use to wrap the go-kit Logger to make it
// adhere to go-migrate's Logger interface.
type logAdapter struct {
	logger  kitlog.Logger
	verbose bool
}

// Printf is semantically the same as fmt.Printf
func (l *logAdapter) Printf(format string, v ...interface{}) {
	l.logger.Log("msg", fmt.Sprintf(format, v...))
}

// Verbose returns true when verbose logging output is wanted
func (l *logAdapter) Verbose() bool {
	return l.verbose
}
