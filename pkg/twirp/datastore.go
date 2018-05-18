package twirp

import (
	"context"

	kitlog "github.com/go-kit/kit/log"
	datastore "github.com/thingful/twirp-datastore-go"

	"github.com/thingful/iotstore/pkg/postgres"
)

type Datastore struct {
	db     *postgres.DB
	logger kitlog.Logger
}

var _ datastore.Datastore = &Datastore{}

func NewDatastore(connStr string, logger kitlog.Logger) *Datastore {
	db := postgres.NewDB(connStr, logger)

	logger = kitlog.With(logger, "module", "twirp")

	ds := &Datastore{
		db:     db,
		logger: logger,
	}

	return ds
}

func (d *Datastore) Start() error {
	return d.db.Start()
}

func (d *Datastore) Stop() error {
	return d.db.Stop()
}

func (d *Datastore) WriteData(ctx context.Context, req *datastore.WriteRequest) (*datastore.WriteResponse, error) {
	return nil, nil
}

func (d *Datastore) ReadData(ctx context.Context, req *datastore.ReadRequest) (*datastore.ReadResponse, error) {
	return nil, nil
}

func (d *Datastore) DeleteData(ctx context.Context, req *datastore.DeleteRequest) (*datastore.DeleteResponse, error) {
	return nil, nil
}
