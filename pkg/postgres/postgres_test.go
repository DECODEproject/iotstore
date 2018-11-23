package postgres_test

import (
	"os"
	"testing"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/DECODEproject/iotstore/pkg/postgres"
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

	postgres.MigrateDownAll(db.DB, logger)

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
	policyId := "abc123"
	deviceToken := "device-token"

	err := s.db.WriteData(policyId, []byte("encrypted bytes"), deviceToken)
	assert.Nil(s.T(), err)

	events, err := s.db.ReadData(policyId, 50, startTime, time.Time{}, "")
	assert.Nil(s.T(), err)
	assert.Len(s.T(), events, 1)

	event := events[0]
	assert.Equal(s.T(), []byte("encrypted bytes"), event.Data)

	err = s.db.DeleteData(time.Now(), false)
	assert.Nil(s.T(), err)

	events, err = s.db.ReadData(policyId, 50, startTime, time.Time{}, "")
	assert.Nil(s.T(), err)
	assert.Len(s.T(), events, 1)

	err = s.db.DeleteData(time.Now(), true)
	assert.Nil(s.T(), err)

	events, err = s.db.ReadData(policyId, 50, startTime, time.Time{}, "")
	assert.Nil(s.T(), err)
	assert.Len(s.T(), events, 0)
}

func (s *PostgresSuite) TestReadWithEndTime() {
	startTime := time.Now().Add(time.Hour * -1)
	endTime := time.Now().Add(time.Minute * -30)
	policyId := "abc123"
	deviceToken := "device-token"

	err := s.db.WriteData(policyId, []byte("encrypted bytes"), deviceToken)
	assert.Nil(s.T(), err)

	events, err := s.db.ReadData(policyId, 50, startTime, endTime, "")
	assert.Nil(s.T(), err)
	assert.Len(s.T(), events, 0)
}

func (s *PostgresSuite) TestPing() {
	err := s.db.Ping()
	assert.Nil(s.T(), err)
}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}
