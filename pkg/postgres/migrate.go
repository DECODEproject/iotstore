package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
)

func MigrateUp(db *sql.DB) error {
	log.Println("Migrating DB up")

	m, err := getMigrator(db)
	if err != nil {
		return err
	}

	return m.Up()
}

func MigrateDown(db *sql.DB) error {
	log.Println("Migrating DB down")

	m, err := getMigrator(db)
	if err != nil {
		return err
	}

	return m.Down()
}

// NewMigration creates a new pair of files into which an SQL migration should
// be written. All this is doing is ensuring files created are correctly named.
func NewMigration(dirName, migrationName string) error {
	if migrationName == "" {
		return errors.New("Must specify a name when creating a migration")
	}

	migrationID := time.Now().Format("20060102150405") + "_" + migrationName
	upFileName := fmt.Sprintf("%s_up.sql", migrationID)
	downFileName := fmt.Sprintf("%s_down.sql", migrationID)

	log.Printf("Creating migration files: '%s' and '%s' in '%s'\n", upFileName, downFileName, dirName)

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

func getMigrator(db *sql.DB) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	return migrate.NewWithDatabaseInstance(
		"file:///pkg/migrations",
		"postgres",
		driver,
	)
}
