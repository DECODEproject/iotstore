package rpc

import (
	"context"
	"fmt"
	"time"

	sq "github.com/elgris/sqrl"
	kitlog "github.com/go-kit/kit/log"
	"github.com/golang/protobuf/ptypes"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	datastore "github.com/thingful/twirp-datastore-go"
	"github.com/twitchtv/twirp"

	"github.com/thingful/iotstore/pkg/postgres"
)

const (
	// DefaultPageSize is the default page size we use when reading data from the
	// datastore.
	DefaultPageSize = 500

	// MaxPageSize is the maximum page size we allow clients to request when
	// reading data from the datastore.
	MaxPageSize = 1000
)

// Datastore is our implementation of the generated twirp interface for the
// encrypted datastore.
type Datastore struct {
	DB      *sqlx.DB
	connStr string
	logger  kitlog.Logger
}

// ensure we adhere to the interface
var _ datastore.Datastore = &Datastore{}

// event is an internal type used when pulling records from the database.
type event struct {
	RecordedAt time.Time `db:"recorded_at"`
	Data       []byte    `db:"data"`
}

// NewDatastore returns a newly instantiated Datastore instance. It takes as
// parameters a DB connection string and a logger. The connection string is
// passed down to the postgres package where it is used to connect.
func NewDatastore(connStr string, logger kitlog.Logger) *Datastore {
	logger = kitlog.With(logger, "module", "rpc")

	ds := &Datastore{
		connStr: connStr,
		logger:  logger,
	}

	return ds
}

// Start starts all child components (here just the postgres DB).
func (d *Datastore) Start() error {
	d.logger.Log("msg", "starting datastore")

	db, err := postgres.Open(d.connStr)
	if err != nil {
		return errors.Wrap(err, "opening db connection failed")
	}

	d.DB = db

	err = postgres.MigrateUp(d.DB.DB, d.logger)
	if err != nil {
		return errors.Wrap(err, "running up migrations failed")
	}

	return nil
}

// Stop stops all child components.
func (d *Datastore) Stop() error {
	d.logger.Log("msg", "stopping datastore")

	return d.DB.Close()
}

// WriteData is the method by which data is written into the datastore. It
// accepts a context, and a WriteRequest object containing the data. Provided
// the incoming request object is valid, then an event will be written into the
// database. Any invalid data will return an error.
func (d *Datastore) WriteData(ctx context.Context, req *datastore.WriteRequest) (*datastore.WriteResponse, error) {
	if req.PublicKey == "" {
		return nil, twirp.RequiredArgumentError("public_key")
	}

	if req.UserUid == "" {
		return nil, twirp.RequiredArgumentError("user_uid")
	}

	sql, args, err := sq.Insert("events").Columns("public_key", "user_uid", "data").Values(req.PublicKey, req.UserUid, req.Data).ToSql()
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	sql = d.DB.Rebind(sql)

	d.logger.Log("sql", sql, "public_key", args[0], "user_uid", args[1])

	_, err = d.DB.Exec(sql, args...)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	return &datastore.WriteResponse{}, nil
}

func (d *Datastore) ReadData(ctx context.Context, req *datastore.ReadRequest) (*datastore.ReadResponse, error) {
	if req.PublicKey == "" {
		return nil, twirp.RequiredArgumentError("public_key")
	}

	if req.PageSize == 0 {
		req.PageSize = DefaultPageSize
	}

	if req.PageSize < 0 || req.PageSize > MaxPageSize {
		return nil, twirp.InvalidArgumentError("page_size", fmt.Sprintf("must be between 1 and %v", MaxPageSize))
	}

	builder := sq.Select("recorded_at", "data").
		From("events").
		OrderBy("recorded_at ASC").
		Where(sq.Eq{"public_key": req.PublicKey}).
		Limit(uint64(req.PageSize))

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	sql = d.DB.Rebind(sql)

	d.logger.Log("msg", "reading events", "sql", sql)

	rows, err := d.DB.Queryx(sql, args...)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	events := []*datastore.EncryptedEvent{}

	for rows.Next() {
		var e event
		err = rows.StructScan(&e)
		if err != nil {
			return nil, twirp.InternalErrorWith(err)
		}

		ev, err := buildEncryptedEvent(&e)
		if err != nil {
			return nil, twirp.InternalErrorWith(err)
		}

		events = append(events, ev)
	}

	return &datastore.ReadResponse{
		PublicKey: req.PublicKey,
		Events:    events,
		PageSize:  req.PageSize,
	}, nil
}

func (d *Datastore) DeleteData(ctx context.Context, req *datastore.DeleteRequest) (*datastore.DeleteResponse, error) {
	if req.UserUid == "" {
		return nil, twirp.RequiredArgumentError("user_uid")
	}

	builder := sq.Delete("").
		From("events").
		Where(sq.Eq{"user_uid": req.UserUid})

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	sql = d.DB.Rebind(sql)

	_, err = d.DB.Exec(sql, args...)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	return &datastore.DeleteResponse{}, nil
}

func buildEncryptedEvent(e *event) (*datastore.EncryptedEvent, error) {
	timestamp, err := ptypes.TimestampProto(e.RecordedAt)
	if err != nil {
		return nil, err
	}

	return &datastore.EncryptedEvent{
		EventTime: timestamp,
		Data:      e.Data,
	}, nil
}
