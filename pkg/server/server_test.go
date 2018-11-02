package server_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	kitlog "github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"

	"github.com/DECODEproject/iotstore/pkg/postgres"
	"github.com/DECODEproject/iotstore/pkg/server"
)

func TestPulseHandler(t *testing.T) {
	connStr := os.Getenv("IOTSTORE_DATABASE_URL")
	logger := kitlog.NewNopLogger()

	db := postgres.NewDB(connStr, true, logger)
	err := db.Start()
	assert.Nil(t, err)

	req, err := http.NewRequest(http.MethodGet, "/pulse", nil)
	assert.Nil(t, err)

	rr := httptest.NewRecorder()
	handler := server.PulseHandler(db)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
