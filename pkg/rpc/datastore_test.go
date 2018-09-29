package rpc_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"

	kitlog "github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	datastore "github.com/thingful/twirp-datastore-go"

	"github.com/thingful/iotstore/pkg/postgres"
	"github.com/thingful/iotstore/pkg/rpc"
)

// getTestDatastore is a helper function that returns a datastore, and also does
// some housekeeping to clean the DB by rolling back and reapplying migrations.
//
// TODO: not terribly happy with this as an approach. See if we can think of an
// alternative.
func getTestDatastore(t *testing.T) *rpc.Datastore {
	t.Helper()

	logger := kitlog.NewNopLogger()
	connStr := os.Getenv("IOTSTORE_DATABASE_URL")

	// create datastore
	ds := rpc.NewDatastore(connStr, true, logger)

	// start the datastore (this runs all migrations slightly annoyingly)
	err := ds.Start()
	if err != nil {
		t.Fatalf("Error starting datastore: %v", err)
	}

	err = postgres.MigrateDownAll(ds.DB.DB, logger)
	if err != nil {
		t.Fatalf("Error running down migrations: %v", err)
	}

	err = postgres.MigrateUp(ds.DB.DB, logger)
	if err != nil {
		t.Fatalf("Error running down migrations: %v", err)
	}

	return ds
}

func TestRoundTrip(t *testing.T) {
	ds := getTestDatastore(t)
	defer ds.Stop()

	startTime, err := ptypes.TimestampProto(time.Now().Add(time.Hour * -1))
	assert.Nil(t, err)

	_, err = ds.WriteData(context.Background(), &datastore.WriteRequest{
		PublicKey: "123abc",
		Data:      []byte("hello world"),
	})
	assert.Nil(t, err)

	var count int
	err = ds.DB.Get(&count, ds.DB.Rebind("SELECT COUNT(*) FROM events WHERE public_key = ?"), "123abc")
	assert.Nil(t, err)
	assert.Equal(t, 1, count)

	resp, err := ds.ReadData(context.Background(), &datastore.ReadRequest{
		PublicKey: "123abc",
		StartTime: startTime,
	})
	assert.Nil(t, err)
	assert.Equal(t, "123abc", resp.PublicKey)
	assert.Len(t, resp.Events, 1)
	assert.Equal(t, int(rpc.DefaultPageSize), int(resp.PageSize))
	assert.Equal(t, "", resp.NextPageCursor)

	event := resp.Events[0]
	assert.Equal(t, []byte("hello world"), event.Data)
}

func TestWriteDataInvalid(t *testing.T) {
	ds := getTestDatastore(t)
	defer ds.Stop()

	testcases := []struct {
		label         string
		request       *datastore.WriteRequest
		expectedError string
	}{
		{
			label:         "missing public_key",
			request:       &datastore.WriteRequest{},
			expectedError: "twirp error invalid_argument: public_key is required",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			_, err := ds.WriteData(context.Background(), tc.request)
			assert.NotNil(t, err)
			assert.Equal(t, tc.expectedError, err.Error())
		})
	}
}

func TestReadDataInvalid(t *testing.T) {
	ds := getTestDatastore(t)
	defer ds.Stop()

	now := time.Now()
	startTime, _ := ptypes.TimestampProto(now)
	invalidEndTime, _ := ptypes.TimestampProto(now.Add(time.Second * -1))

	testcases := []struct {
		label         string
		request       *datastore.ReadRequest
		expectedError string
	}{
		{
			label: "missing public_key",
			request: &datastore.ReadRequest{
				StartTime: startTime,
			},
			expectedError: "twirp error invalid_argument: public_key is required",
		},
		{
			label: "large page size",
			request: &datastore.ReadRequest{
				PublicKey: "123abc",
				StartTime: startTime,
				PageSize:  1001,
			},
			expectedError: "twirp error invalid_argument: page_size must be between 1 and 1000",
		},
		{
			label: "missing start time",
			request: &datastore.ReadRequest{
				PublicKey: "123abc",
			},
			expectedError: "twirp error invalid_argument: start_time is required",
		},
		{
			label: "end_time before start_time",
			request: &datastore.ReadRequest{
				PublicKey: "123abc",
				StartTime: startTime,
				EndTime:   invalidEndTime,
			},
			expectedError: "twirp error invalid_argument: end_time must be after start_time",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			_, err := ds.ReadData(context.Background(), tc.request)
			assert.NotNil(t, err)
			assert.Equal(t, tc.expectedError, err.Error())
		})
	}
}

func TestPagination(t *testing.T) {
	ds := getTestDatastore(t)
	defer ds.Stop()

	startTime, _ := time.Parse(time.RFC3339, "2018-05-01T08:00:00Z")
	endTime, _ := time.Parse(time.RFC3339, "2018-05-01T08:03:00Z")

	startTimestamp, _ := ptypes.TimestampProto(startTime)
	endTimestamp, _ := ptypes.TimestampProto(endTime)

	fixtures := []struct {
		publicKey string
		timestamp string
		data      []byte
	}{
		{
			publicKey: "abc123",
			timestamp: "2018-05-01T08:00:00Z",
			data:      []byte("first"),
		},
		{
			publicKey: "abc123",
			timestamp: "2018-05-01T08:02:00Z",
			data:      []byte("third"),
		},
		{
			publicKey: "abc123",
			timestamp: "2018-05-01T08:01:00Z",
			data:      []byte("second"),
		},
		{
			publicKey: "abc123",
			timestamp: "2018-05-01T08:02:00Z",
			data:      []byte("fourth"),
		},
	}

	// load fixtures into db
	for _, f := range fixtures {
		ts, _ := time.Parse(time.RFC3339, f.timestamp)

		ds.DB.MustExec("INSERT INTO events (public_key, recorded_at, data) VALUES ($1, $2, $3)", f.publicKey, ts, f.data)
	}

	resp, err := ds.ReadData(context.Background(), &datastore.ReadRequest{
		PublicKey: "abc123",
		PageSize:  3,
		StartTime: startTimestamp,
		EndTime:   endTimestamp,
	})

	assert.Nil(t, err)
	assert.Equal(t, "abc123", resp.PublicKey)
	assert.Len(t, resp.Events, 3)
	assert.NotEqual(t, "", resp.NextPageCursor)

	assert.Equal(t, "first", string(resp.Events[0].Data))
	assert.Equal(t, "second", string(resp.Events[1].Data))
	assert.Equal(t, "third", string(resp.Events[2].Data))

	resp, err = ds.ReadData(context.Background(), &datastore.ReadRequest{
		PublicKey:  "abc123",
		PageSize:   3,
		PageCursor: resp.NextPageCursor,
		StartTime:  startTimestamp,
		EndTime:    endTimestamp,
	})

	assert.Nil(t, err)
	assert.Len(t, resp.Events, 1)
	assert.Equal(t, "", resp.NextPageCursor)
}
