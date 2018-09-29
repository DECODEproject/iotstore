package rpc

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
	verbose bool
}

// ensure we adhere to the interface
var _ datastore.Datastore = &Datastore{}

// event is an internal type used when pulling records from the database.
type event struct {
	ID         int64     `db:"id"`
	RecordedAt time.Time `db:"recorded_at"`
	Data       []byte    `db:"data"`
}

// cursor is an internal type used when paginating.
type cursor struct {
	EventID   int64     `json:"eventID"`
	Timestamp time.Time `json:"timestamp"`
}

// NewDatastore returns a newly instantiated Datastore instance. It takes as
// parameters a DB connection string and a logger. The connection string is
// passed down to the postgres package where it is used to connect.
func NewDatastore(connStr string, verbose bool, logger kitlog.Logger) *Datastore {
	logger = kitlog.With(logger, "module", "rpc")

	ds := &Datastore{
		connStr: connStr,
		logger:  logger,
		verbose: verbose,
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

	if d.verbose {
		d.logger.Log(
			"publicKey", req.PublicKey,
			"msg", "WriteData",
			"encodedPayload", string(req.Data),
		)
	}

	sql, args, err := sq.Insert("events").Columns("public_key", "data").
		Values(req.PublicKey, req.Data).ToSql()

	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	sql = d.DB.Rebind(sql)

	_, err = d.DB.Exec(sql, args...)
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
			"publicKey", req.PublicKey,
			"pageSize", req.PageSize,
			"msg", "ReadData",
		)
	}

	builder := sq.Select("id", "recorded_at", "data").
		From("events").
		OrderBy("recorded_at ASC", "id ASC").
		Where(sq.Eq{"public_key": req.PublicKey}).
		Where(sq.GtOrEq{"recorded_at": startTime}).
		Limit(uint64(req.PageSize))

	if !endTime.IsZero() {
		builder = builder.Where(sq.Lt{"recorded_at": endTime})
	}

	if req.PageCursor != "" {
		// decode the cursor
		c, err := decodeCursor(req.PageCursor)
		if err != nil {
			return nil, twirp.InternalErrorWith(err)
		}

		builder = builder.Where(sq.GtOrEq{"recorded_at": c.Timestamp}).Where(sq.Gt{"id": c.EventID})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	sql = d.DB.Rebind(sql)

	rows, err := d.DB.Queryx(sql, args...)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	events := []*datastore.EncryptedEvent{}
	var lastEventID int64

	for rows.Next() {
		var e event
		err = rows.StructScan(&e)
		if err != nil {
			return nil, twirp.InternalErrorWith(err)
		}

		lastEventID = e.ID

		ev, err := buildEncryptedEvent(&e)
		if err != nil {
			return nil, twirp.InternalErrorWith(err)
		}

		events = append(events, ev)
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
	c := &cursor{
		Timestamp: timestamp,
		EventID:   lastEventID,
	}

	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

// decodeCursor is a helper function that takes our cursor string if set, then
// reverses the encoding process which here is decoding the base64 string, and
// then parsing the JSON into our cursor type.
func decodeCursor(in string) (*cursor, error) {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return nil, err
	}

	var c cursor
	err = json.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
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
