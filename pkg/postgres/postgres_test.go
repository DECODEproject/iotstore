package postgres_test

import (
	"context"
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

func (s *PostgresSuite) TestRoundTripEvent() {
	startTime := time.Now().Add(time.Hour * -1)
	policyId := "abc123"
	deviceToken := "device-token"

	err := s.db.WriteData(policyId, []byte("encrypted bytes"), deviceToken)
	assert.Nil(s.T(), err)
	err = s.db.WriteData(policyId, []byte("encrypted bytes"), deviceToken)
	assert.Nil(s.T(), err)
	err = s.db.WriteData(policyId, []byte("encrypted bytes"), deviceToken)
	assert.Nil(s.T(), err)
	err = s.db.WriteData(policyId, []byte("encrypted bytes"), deviceToken)
	assert.Nil(s.T(), err)

	page, err := s.db.ReadData(policyId, 3, startTime, time.Time{}, "")
	assert.Nil(s.T(), err)
	assert.Len(s.T(), page.Events, 3)
	assert.NotEqual(s.T(), "", page.NextPageCursor)

	event := page.Events[0]
	assert.Equal(s.T(), []byte("encrypted bytes"), event.Data)

	// get next page
	page, err = s.db.ReadData(policyId, 3, startTime, time.Time{}, page.NextPageCursor)
	assert.Nil(s.T(), err)
	assert.Len(s.T(), page.Events, 1)

	err = s.db.DeleteData(time.Now(), false)
	assert.Nil(s.T(), err)

	page, err = s.db.ReadData(policyId, 50, startTime, time.Time{}, "")
	assert.Nil(s.T(), err)
	assert.Len(s.T(), page.Events, 4)

	err = s.db.DeleteData(time.Now(), true)
	assert.Nil(s.T(), err)

	page, err = s.db.ReadData(policyId, 50, startTime, time.Time{}, "")
	assert.Nil(s.T(), err)
	assert.Len(s.T(), page.Events, 0)
}

func (s *PostgresSuite) TestReadWithEndTime() {
	startTime := time.Now().Add(time.Hour * -1)
	endTime := time.Now().Add(time.Minute * -30)
	policyId := "abc123"
	deviceToken := "device-token"

	err := s.db.WriteData(policyId, []byte("encrypted bytes"), deviceToken)
	assert.Nil(s.T(), err)

	page, err := s.db.ReadData(policyId, 50, startTime, endTime, "")
	assert.Nil(s.T(), err)
	assert.Len(s.T(), page.Events, 0)
}

func (s *PostgresSuite) TestPing() {
	err := s.db.Ping()
	assert.Nil(s.T(), err)
}

func (s *PostgresSuite) TestCertificates() {
	ctx := context.Background()

	// nonexistent key should return error
	_, err := s.db.Get(ctx, "baz")
	assert.NotNil(s.T(), err)

	// should be able to write a cert
	err = s.db.Put(ctx, "foo", []byte("bar"))
	assert.Nil(s.T(), err)

	// now should be able to read it
	cert, err := s.db.Get(ctx, "foo")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []byte("bar"), cert)

	// should be able to delete it
	err = s.db.Delete(ctx, "foo")
	assert.Nil(s.T(), err)

	// now should not be able to read it
	_, err = s.db.Get(ctx, "foo")
	assert.NotNil(s.T(), err)
}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}
