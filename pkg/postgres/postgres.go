package postgres

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	sq "github.com/elgris/sqrl"
	raven "github.com/getsentry/raven-go"
	kitlog "github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // required way of importing pq
	"github.com/pkg/errors"
)

// Open takes as input a connection string for a DB, and returns either a
// sqlx.DB instance or an error.
func Open(connStr string) (*sqlx.DB, error) {
	return sqlx.Open("postgres", connStr)
}

// Event is a type used to read encrypted events back from the database
type Event struct {
	ID         int64     `db:"id"`
	RecordedAt time.Time `db:"recorded_at"`
	Data       []byte    `db:"data"`
}

// Cursor is an internal type used for serializing or parsing page cursors.
type Cursor struct {
	EventID   int64     `json:"eventID"`
	Timestamp time.Time `json:"timestamp"`
}

// Page is a struct used to return a page of events. Includes a cursor which
// points to the next page rather than have the twirp layer do this work.
type Page struct {
	Events         []*Event
	NextPageCursor string
}

// Certificate is an internal type used for persisting and reading TLS
// certificates from Postgres
type Certificate struct {
	Key         string `db:"key"`
	Certificate []byte `db:"certificate"`
}

// DB is a struct that wraps an sqlx.DB instance that exposes some methods to
// read and write data.
type DB struct {
	DB *sqlx.DB

	connStr string
	verbose bool
	logger  kitlog.Logger
}

// NewDB is a constructor that returns a new DB instance for the given
// configuration. We pass in a connection string for the database, and in
// addition we can pass in a verbose flag which causes the binary to output more
// verbose log information.
func NewDB(connStr string, verbose bool, logger kitlog.Logger) *DB {
	logger = kitlog.With(logger, "module", "postgres")

	db := &DB{
		connStr: connStr,
		verbose: verbose,
		logger:  logger,
	}

	return db
}

// Start connects to the database, and runs any pending up migrations.
func (d *DB) Start() error {
	d.logger.Log("msg", "starting postgres connection")

	db, err := Open(d.connStr)
	if err != nil {
		return errors.Wrap(err, "failed to open db connection")
	}

	d.DB = db

	err = MigrateUp(d.DB.DB, d.logger)
	if err != nil {
		return errors.Wrap(err, "failed to run up migrations")
	}

	return nil
}

// Stop terminates the DB connection, closing the pool of connections.
func (d *DB) Stop() error {
	d.logger.Log("msg", "stopping postgres connection")

	return d.DB.Close()
}

// WriteData is the function that is responsible for writing data to the actual
// database. Takes as input the id of the policy the policy for which we are
// storing data and a byte slice containing the encrypted data to be persisted.
// In addition we also pass in the unique device token.
func (d *DB) WriteData(policyId string, data []byte, deviceToken string) error {
	sql := `INSERT INTO events
		(policy_id, data, device_token)
		VALUES (:policy_id, :data, :device_token)`

	mapArgs := map[string]interface{}{
		"policy_id":    policyId,
		"data":         data,
		"device_token": deviceToken,
	}

	tx, err := d.DB.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	sql, args, err := tx.BindNamed(sql, mapArgs)
	if err != nil {
		tx.Rollback()
		raven.CaptureError(err, map[string]string{"operation": "writeData"})
		return errors.Wrap(err, "failed to bind named query")
	}

	_, err = tx.Exec(sql, args...)
	if err != nil {
		tx.Rollback()
		raven.CaptureError(err, map[string]string{"operation": "writeData"})
		return errors.Wrap(err, "failed to execute write query")
	}

	return tx.Commit()
}

// ReadData returns a list of Event types for the given query parameters.
func (d *DB) ReadData(policyId string, pageSize uint64, startTime, endTime time.Time, pageCursor string) (*Page, error) {
	// use sqrl builder here as it simplifies the creation of the query.
	builder := sq.Select("id", "recorded_at", "data").
		From("events").
		OrderBy("recorded_at ASC", "id ASC").
		Where(sq.Eq{"policy_id": policyId}).
		Where(sq.GtOrEq{"recorded_at": startTime}).
		Limit(pageSize)

	if !endTime.IsZero() {
		builder = builder.Where(sq.Lt{"recorded_at": endTime})
	}

	if pageCursor != "" {
		cursor, err := decodeCursor(pageCursor)
		if err != nil {
			raven.CaptureError(err, map[string]string{"operation": "readData"})
			return nil, errors.Wrap(err, "failed to decode page cursor")
		}
		builder = builder.Where(sq.GtOrEq{"recorded_at": cursor.Timestamp}).Where(sq.Gt{"id": cursor.EventID})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		raven.CaptureError(err, map[string]string{"operation": "readData"})
		return nil, errors.Wrap(err, "failed to build sql query")
	}

	sql = d.DB.Rebind(sql)

	rows, err := d.DB.Queryx(sql, args...)
	if err != nil {
		raven.CaptureError(err, map[string]string{"operation": "readData"})
		return nil, errors.Wrap(err, "failed to execute query")
	}

	events := []*Event{}

	for rows.Next() {
		var e Event
		err = rows.StructScan(&e)
		if err != nil {
			raven.CaptureError(err, map[string]string{"operation": "readData"})
			return nil, errors.Wrap(err, "failed to populate Event type")
		}
		events = append(events, &e)
	}

	var nextCursor string

	if len(events) == int(pageSize) {
		// we should construct a next page cursor value to return
		nextCursor, err = encodeCursor(events[len(events)-1])
		if err != nil {
			raven.CaptureError(err, map[string]string{"operation": "readData"})
			return nil, errors.Wrap(err, "failed to build next page cursor")
		}
	}

	return &Page{
		Events:         events,
		NextPageCursor: nextCursor,
	}, nil
}

// DeleteData takes as input a timestamp, and after it is executed will have
// deleted all events in the database from before the submitted time interval
// for any policy. This function also takes a `execute` parameter. If set to
// true the delete operation is performed and committed, but if set to false it
// is executed without committing the transaction. This allows a caller to see
// how many events would be deleted.
func (d *DB) DeleteData(before time.Time, execute bool) error {
	sql := `WITH deleted AS
		(DELETE FROM events WHERE recorded_at < ? RETURNING *)
		SELECT COUNT(*) FROM deleted`

	sql = d.DB.Rebind(sql)

	if d.verbose {
		d.logger.Log(
			"msg", "deleting old events",
			"sql", sql,
			"before", before.Format(time.RFC3339),
			"execute", execute,
		)
	}

	tx, err := d.DB.Beginx()
	if err != nil {
		raven.CaptureError(err, map[string]string{"operation": "deleteData"})
		return errors.Wrap(err, "failed to start transaction")
	}

	var count int
	err = tx.Get(&count, sql, before)
	if err != nil {
		tx.Rollback()
		raven.CaptureError(err, map[string]string{"operation": "deleteData"})
		return errors.Wrap(err, "failed to execute delete query")
	}

	d.logger.Log("msg", "deleted old events", "count", count, "execute", execute)

	if !execute {
		return tx.Rollback()
	}

	return tx.Commit()
}

// Ping attempts to verify a connection to the database is still alive,
// establishing a connection if necessary by executing a simple select query
// agianst the DB. Note using DB.Ping() did not work as expected as if there are
// existing connections in the pool that aren't used it will return no error
// without actually going to the DB to check.
func (d *DB) Ping() error {
	_, err := d.DB.Exec("SELECT 1")
	if err != nil {
		return err
	}
	return nil
}

// Get is our implementation of the method defined in the autocert.Cache
// interface for reading certificates from some underlying datastore.
func (d *DB) Get(ctx context.Context, key string) ([]byte, error) {
	sql := `SELECT certificate FROM certificates WHERE key = $1`

	var certificate []byte
	err := d.DB.Get(&certificate, sql, key)
	if err != nil {
		raven.CaptureError(err, map[string]string{"operation": "getCertificate"})
		return nil, errors.Wrap(err, "failed to read certificate from DB")
	}

	return certificate, nil
}

// Put is our implementation of the method defined in the autocert.Cache
// interface for writing certificates to some underlying datastore.
func (d *DB) Put(ctx context.Context, key string, data []byte) error {
	sql := `INSERT INTO certificates (key, certificate)
		VALUES (:key, :certificate)
	ON CONFLICT (key)
	DO UPDATE SET certificate = EXCLUDED.certificate`

	mapArgs := map[string]interface{}{
		"key":         key,
		"certificate": data,
	}

	tx, err := d.DB.Beginx()
	if err != nil {
		raven.CaptureError(err, map[string]string{"operation": "putCertificate"})
		return errors.Wrap(err, "failed to begin transaction when writing certificate")
	}

	sql, args, err := tx.BindNamed(sql, mapArgs)
	if err != nil {
		tx.Rollback()
		raven.CaptureError(err, map[string]string{"operation": "putCertificate"})
		return errors.Wrap(err, "failed to bind named parameters")
	}

	_, err = tx.Exec(sql, args...)
	if err != nil {
		tx.Rollback()
		raven.CaptureError(err, map[string]string{"operation": "putCertificate"})
		return errors.Wrap(err, "failed to insert certificate")
	}

	return tx.Commit()
}

// Delete is our implementation of the autocert.Cache interface for deleting
// certificates from some underlying datastore.
func (d *DB) Delete(ctx context.Context, key string) error {
	sql := `DELETE FROM certificates WHERE key = $1`

	tx, err := d.DB.Beginx()
	if err != nil {
		raven.CaptureError(err, map[string]string{"operation": "deleteCertificate"})
		return errors.Wrap(err, "failed to begin transaction when deleting certificate")
	}

	_, err = tx.Exec(sql, key)
	if err != nil {
		tx.Rollback()
		raven.CaptureError(err, map[string]string{"operation": "deleteCertificate"})
		return errors.Wrap(err, "failed to delete certificate")
	}

	return tx.Commit()
}

func encodeCursor(event *Event) (string, error) {
	// create non-empty cursor meaning the requestor can look for more pages
	c := &Cursor{
		Timestamp: event.RecordedAt,
		EventID:   event.ID,
	}

	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

// decodeCursor takes as input an encoded cursor string. Note we could have used
// protobuffers here too, but for now this is just a simple struct encoded to
// JSON and base64 encoded.
func decodeCursor(in string) (*Cursor, error) {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return nil, err
	}

	var c Cursor
	err = json.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
