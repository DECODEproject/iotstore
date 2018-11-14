package rpc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/golang/protobuf/ptypes"
	datastore "github.com/thingful/twirp-datastore-go"
	"github.com/twitchtv/twirp"

	"github.com/DECODEproject/iotstore/pkg/postgres"
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
	DB      *postgres.DB
	logger  kitlog.Logger
	verbose bool
}

// ensure we adhere to the interface
var _ datastore.Datastore = &Datastore{}

// NewDatastore returns a newly instantiated Datastore instance. It takes as
// parameters a DB connection string and a logger. The connection string is
// passed down to the postgres package where it is used to connect.
func NewDatastore(connStr string, verbose bool, logger kitlog.Logger) *Datastore {
	db := postgres.NewDB(connStr, verbose, logger)

	logger = kitlog.With(logger, "module", "rpc")

	ds := &Datastore{
		DB:      db,
		logger:  logger,
		verbose: verbose,
	}

	return ds
}

// Start starts all child components (here just the postgres DB).
func (d *Datastore) Start() error {
	d.logger.Log("msg", "starting datastore")

	return d.DB.Start()
}

// Stop stops all child components.
func (d *Datastore) Stop() error {
	d.logger.Log("msg", "stopping datastore")

	return d.DB.Stop()
}

// WriteData is the method by which data is written into the datastore. It
// accepts a context, and a WriteRequest object containing the data. Provided
// the incoming request object is valid, then an event will be written into the
// database. Any invalid data will return an error.
func (d *Datastore) WriteData(ctx context.Context, req *datastore.WriteRequest) (*datastore.WriteResponse, error) {
	if req.PublicKey == "" {
		return nil, twirp.RequiredArgumentError("public_key")
	}

	if req.DeviceToken == "" {
		return nil, twirp.RequiredArgumentError("device_token")
	}

	if d.verbose {
		d.logger.Log(
			"publicKey", req.PublicKey,
			"msg", "WriteData",
			"encodedPayload", string(req.Data),
		)
	}

	err := d.DB.WriteData(req.PublicKey, req.Data, req.DeviceToken)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	return &datastore.WriteResponse{}, nil
}

// ReadData is the handler that allows a client to request data from the
// datastore matching search parameters defined in the incoming
// datastore.ReadRequest object.
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

	startTime, endTime, err := extractTimes(req)
	if err != nil {
		return nil, err
	}

	if d.verbose {
		d.logger.Log(
			"msg", "ReadData",
			"publicKey", req.PublicKey,
			"pageSize", req.PageSize,
			"startTime", startTime,
			"endTime", endTime,
		)
	}

	rawEvents, err := d.DB.ReadData(req.PublicKey, uint64(req.PageSize), startTime, endTime, req.PageCursor)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	events := []*datastore.EncryptedEvent{}
	var lastEventID int64

	for _, raw := range rawEvents {
		lastEventID = raw.ID

		event, err := buildEncryptedEvent(raw)
		if err != nil {
			return nil, twirp.InternalErrorWith(err)
		}

		events = append(events, event)
	}

	nextCursor, err := encodeCursor(events, req.PageSize, lastEventID)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	return &datastore.ReadResponse{
		PublicKey:      req.PublicKey,
		Events:         events,
		PageSize:       req.PageSize,
		NextPageCursor: nextCursor,
	}, nil
}

// buildEncryptedEvent is a helper function that converts our internal event
// type read from the database into an external datastore.EncryptedEvent.
func buildEncryptedEvent(e *postgres.Event) (*datastore.EncryptedEvent, error) {
	timestamp, err := ptypes.TimestampProto(e.RecordedAt)
	if err != nil {
		return nil, err
	}

	return &datastore.EncryptedEvent{
		EventTime: timestamp,
		Data:      e.Data,
	}, nil
}

// encodeCursor returns either a marshalled cursor instance if there may be
// more results to fetch (i.e. the number of events equals the page size), an
// empty string if no results are possible (length of results is less than the
// page size), or an error should we fail to generate the new cursor in any way.
func encodeCursor(events []*datastore.EncryptedEvent, pageSize uint32, lastEventID int64) (string, error) {
	if len(events) < int(pageSize) {
		return "", nil
	}
	// there might be more so build a cursor based on the last event
	lastEvent := events[len(events)-1]

	// convert timestamp to time.Time
	timestamp, err := ptypes.Timestamp(lastEvent.EventTime)
	if err != nil {
		return "", err
	}

	// create non-empty cursor meaning the requestor can look for more pages
	c := &postgres.Cursor{
		Timestamp: timestamp,
		EventID:   lastEventID,
	}

	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

// extractTimes extracts the start and end times from an incoming request,
// converts the protobuf Timestamp instances into vanilla time.Time instances.
// We return errors in the following cases: we are unable to convert either
// timestamp into a time.Time, start_time is nil, end_time is before start_time.
func extractTimes(req *datastore.ReadRequest) (startTime time.Time, endTime time.Time, err error) {
	if req.StartTime == nil {
		return startTime, endTime, twirp.RequiredArgumentError("start_time")
	}

	startTime, err = ptypes.Timestamp(req.StartTime)
	if err != nil {
		return startTime, endTime, twirp.InternalErrorWith(err)
	}

	if req.EndTime != nil {
		endTime, err = ptypes.Timestamp(req.EndTime)
		if err != nil {
			return startTime, endTime, twirp.InternalErrorWith(err)
		}

		if endTime.Before(startTime) {
			return startTime, endTime, twirp.InvalidArgumentError("end_time", "must be after start_time")
		}
	}

	return startTime, endTime, nil
}
