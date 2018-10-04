package postgres_test

import (
	"os"
	"testing"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/thingful/iotstore/pkg/postgres"
)

type PostgresSuite struct {
	suite.Suite
	db *postgres.DB
}

func (s *PostgresSuite) SetupTest() {
	logger := kitlog.NewNopLogger()
	connStr := os.Getenv("IOTSTORE_DATABASE_URL")

	db, err := postgres.Open(connStr)
	if err != nil {
		s.T().Fatalf("Failed to open db connection: %v", err)
	}

	err = postgres.MigrateDownAll(db.DB, logger)
	if err != nil {
		s.T().Fatalf("Failed to run down migrations: %v", err)
	}

	err = db.Close()
	if err != nil {
		s.T().Fatalf("Failed to close DB: %v", err)
	}

	s.db = postgres.NewDB(connStr, true, logger)

	err = s.db.Start()
	if err != nil {
		s.T().Fatalf("Failed to start component: %v", err)
	}
}

func (s *PostgresSuite) TearDownTest() {
	s.db.Stop()
}

func (s *PostgresSuite) TestRoundTrip() {
	startTime := time.Now().Add(time.Hour * -1)
	publicKey := "abc123"

	err := s.db.WriteData(publicKey, []byte("encrypted bytes"))
	assert.Nil(s.T(), err)

	events, err := s.db.ReadData(publicKey, 50, startTime, time.Time{}, "")
	assert.Nil(s.T(), err)
	assert.Len(s.T(), events, 1)

	event := events[0]
	assert.Equal(s.T(), []byte("encrypted bytes"), event.Data)

	err = s.db.DeleteData(time.Now(), false)
	assert.Nil(s.T(), err)

	events, err = s.db.ReadData(publicKey, 50, startTime, time.Time{}, "")
	assert.Nil(s.T(), err)
	assert.Len(s.T(), events, 1)

	err = s.db.DeleteData(time.Now(), true)
	assert.Nil(s.T(), err)

	events, err = s.db.ReadData(publicKey, 50, startTime, time.Time{}, "")
	assert.Nil(s.T(), err)
	assert.Len(s.T(), events, 0)
}

func (s *PostgresSuite) TestReadWithEndTime() {
	startTime := time.Now().Add(time.Hour * -1)
	endTime := time.Now().Add(time.Minute * -30)
	publicKey := "abc123"

	err := s.db.WriteData(publicKey, []byte("encrypted bytes"))
	assert.Nil(s.T(), err)

	events, err := s.db.ReadData(publicKey, 50, startTime, endTime, "")
	assert.Nil(s.T(), err)
	assert.Len(s.T(), events, 0)
}

//func (s *PostgresSuite) TestPagination() {
//	//startTime, _ := time.Parse(time.RFC3339, "2018-05-01T08:00:00Z")
//	//endTime, _ := time.Parse(time.RFC3339, "2018-05-01T08:03:00Z")
//	publicKey := "abc123"
//
//	fixtures := []struct {
//		publicKey string
//		timestamp string
//		data      []byte
//	}{
//		{
//			publicKey: publicKey,
//			timestamp: "2018-05-01T07:59:59",
//			data:      []byte("first"),
//		},
//		{
//			publicKey: publicKey,
//			timestamp: "2018-05-01T08:00:00Z",
//			data:      []byte("second"),
//		},
//		{
//			publicKey: publicKey,
//			timestamp: "2018-05-01T08:02:00Z",
//			data:      []byte("fourth"),
//		},
//		{
//			publicKey: publicKey,
//			timestamp: "2018-05-01T08:01:00Z",
//			data:      []byte("third"),
//		},
//		{
//			publicKey: publicKey,
//			timestamp: "2018-05-01T08:02:00Z",
//			data:      []byte("fourth"),
//		},
//		{
//			publicKey: publicKey,
//			timestamp: "2018-05-01T08:04:00Z",
//			data:      []byte("fifth"),
//		},
//	}
//
//	// load fixtures into db
//	for _, f := range fixtures {
//		ts, _ := time.Parse(time.RFC3339, f.timestamp)
//
//		s.db.DB.MustExec("INSERT INTO events (public_key, recorded_at, data) VALUES ($1, $2, $3)", f.publicKey, ts, f.data)
//	}
//}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}
