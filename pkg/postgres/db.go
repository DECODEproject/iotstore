package postgres

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type DB struct {
	connStr string
	db      *sql.DB
}

func NewDB(connStr string) *DB {
	return &DB{
		connStr: connStr,
	}
}

func (d *DB) Start() error {
	log.Println("Starting DB")

	db, err := sql.Open("postgres", d.connStr)
	if err != nil {
		return err
	}

	d.db = db

	return nil
}

func (d *DB) Stop() error {
	log.Println("Stopping DB")

	return d.db.Close()
}
