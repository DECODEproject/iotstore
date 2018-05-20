package rpc

import (
	"context"

	"github.com/twitchtv/twirp"

	kitlog "github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	datastore "github.com/thingful/twirp-datastore-go"

	"github.com/thingful/iotstore/pkg/postgres"
)

// Datastore is our implementation of the generated twirp interface for the
// encrypted datastore.
type Datastore struct {
	connStr string
	db      *sqlx.DB
	logger  kitlog.Logger
}

// ensure we adhere to the interface
var _ datastore.Datastore = &Datastore{}

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

	d.db = db

	err = postgres.MigrateUp(d.db.DB, d.logger)
	if err != nil {
		return errors.Wrap(err, "running up migrations failed")
	}

	return nil
}

// Stop stops all child components.
func (d *Datastore) Stop() error {
	d.logger.Log("msg", "stopping datastore")

	return d.db.Close()
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

	return nil, nil
}

func (d *Datastore) ReadData(ctx context.Context, req *datastore.ReadRequest) (*datastore.ReadResponse, error) {
	return nil, nil
}

func (d *Datastore) DeleteData(ctx context.Context, req *datastore.DeleteRequest) (*datastore.DeleteResponse, error) {
	return nil, nil
}
