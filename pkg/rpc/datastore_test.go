package rpc_test

import (
	"context"
	"os"
	"testing"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/DECODEproject/iotstore/pkg/postgres"
	"github.com/DECODEproject/iotstore/pkg/rpc"
	datastore "github.com/thingful/twirp-datastore-go"
)

type DatastoreSuite struct {
	suite.Suite
	ds *rpc.Datastore
}

func (s *DatastoreSuite) SetupTest() {
	logger := kitlog.NewNopLogger()
	connStr := os.Getenv("IOTSTORE_DATABASE_URL")

	db, err := postgres.Open(connStr)
	if err != nil {
		s.T().Fatalf("Failed to open db connection: %v", err)
	}

	postgres.MigrateDownAll(db.DB, logger)

	err = db.Close()
	if err != nil {
		s.T().Fatalf("Failed to close DB: %v", err)
	}

	s.ds = rpc.NewDatastore(connStr, true, logger)

	err = s.ds.Start()
	if err != nil {
		s.T().Fatalf("Failed to start datsatore: %v", err)
	}
}

func (s *DatastoreSuite) TearDownTest() {
	err := s.ds.Stop()
	if err != nil {
		s.T().Fatalf("Failed to stop datastore: %v", err)
	}
}

func (s *DatastoreSuite) TestRoundTrip() {
	publicKey := "abc123"
	deviceToken := "device-token"

	startTime, err := ptypes.TimestampProto(time.Now().Add(time.Hour * -1))
	assert.Nil(s.T(), err)

	_, err = s.ds.WriteData(context.Background(), &datastore.WriteRequest{
		PublicKey:   publicKey,
		Data:        []byte("hello world"),
		DeviceToken: deviceToken,
	})
	assert.Nil(s.T(), err)

	resp, err := s.ds.ReadData(context.Background(), &datastore.ReadRequest{
		PublicKey: publicKey,
		StartTime: startTime,
	})

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), publicKey, resp.PublicKey)
	assert.Len(s.T(), resp.Events, 1)
	assert.Equal(s.T(), int(rpc.DefaultPageSize), int(resp.PageSize))
	assert.Equal(s.T(), "", resp.NextPageCursor)

	event := resp.Events[0]
	assert.Equal(s.T(), []byte("hello world"), event.Data)
}

func (s *DatastoreSuite) TestWriteDataInvalid() {
	testcases := []struct {
		label         string
		request       *datastore.WriteRequest
		expectedError string
	}{
		{
			label:         "missing public_key",
			request:       &datastore.WriteRequest{DeviceToken: "device-token"},
			expectedError: "twirp error invalid_argument: public_key is required",
		},
		{
			label:         "missing device_token",
			request:       &datastore.WriteRequest{PublicKey: "abc123"},
			expectedError: "twirp error invalid_argument: device_token is required",
		},
	}

	for _, tc := range testcases {
		s.T().Run(tc.label, func(t *testing.T) {
			_, err := s.ds.WriteData(context.Background(), tc.request)
			assert.NotNil(t, err)
			assert.Equal(t, tc.expectedError, err.Error())
		})
	}
}

func (s *DatastoreSuite) TestReadDataInvalid() {
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
		s.T().Run(tc.label, func(t *testing.T) {
			_, err := s.ds.ReadData(context.Background(), tc.request)
			assert.NotNil(t, err)
			assert.Equal(t, tc.expectedError, err.Error())
		})
	}
}

func (s *DatastoreSuite) TestPagination() {
	startTime, _ := time.Parse(time.RFC3339, "2018-05-01T08:00:00Z")
	endTime, _ := time.Parse(time.RFC3339, "2018-05-01T08:03:00Z")

	startTimestamp, _ := ptypes.TimestampProto(startTime)
	endTimestamp, _ := ptypes.TimestampProto(endTime)

	deviceToken := "device-token"

	fixtures := []struct {
		publicKey string
		timestamp string
		data      []byte
	}{
		{
			publicKey: "abc123",
			timestamp: "2018-05-01T07:59:59",
			data:      []byte("first"),
		},
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
		{
			publicKey: "abc123",
			timestamp: "2018-05-01T08:04:00Z",
			data:      []byte("fourth"),
		},
	}

	// load fixtures into db
	for _, f := range fixtures {
		ts, _ := time.Parse(time.RFC3339, f.timestamp)

		s.ds.DB.DB.MustExec("INSERT INTO events (public_key, recorded_at, data, device_token) VALUES ($1, $2, $3, $4)", f.publicKey, ts, f.data, deviceToken)
	}

	resp, err := s.ds.ReadData(context.Background(), &datastore.ReadRequest{
		PublicKey: "abc123",
		PageSize:  3,
		StartTime: startTimestamp,
		EndTime:   endTimestamp,
	})

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "abc123", resp.PublicKey)
	assert.Len(s.T(), resp.Events, 3)
	assert.NotEqual(s.T(), "", resp.NextPageCursor)

	assert.Equal(s.T(), "first", string(resp.Events[0].Data))
	assert.Equal(s.T(), "second", string(resp.Events[1].Data))
	assert.Equal(s.T(), "third", string(resp.Events[2].Data))

	resp, err = s.ds.ReadData(context.Background(), &datastore.ReadRequest{
		PublicKey:  "abc123",
		PageSize:   3,
		PageCursor: resp.NextPageCursor,
		StartTime:  startTimestamp,
		EndTime:    endTimestamp,
	})

	assert.Nil(s.T(), err)
	assert.Len(s.T(), resp.Events, 1)
	assert.Equal(s.T(), "", resp.NextPageCursor)
}

func TestDatastoreSuite(t *testing.T) {
	suite.Run(t, new(DatastoreSuite))
}
