package server_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"

	"github.com/thingful/iotstore/pkg/server"
)

func TestPulseHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/pulse", nil)
	assert.Nil(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.PulseHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestStartStop(t *testing.T) {
	logger := kitlog.NewNopLogger()
	s := server.NewServer("127.0.0.1:0", os.Getenv("IOTSTORE_DATABASE_URL"), logger)

	go func() {
		s.Start()
	}()

	time.Sleep(time.Second * 1)

	err := s.Stop()
	assert.Nil(t, err)
}
