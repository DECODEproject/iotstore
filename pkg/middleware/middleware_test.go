package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thingful/iotstore/pkg/middleware"
)

func testHandler() http.HandlerFunc {
	fn := func(rw http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(rw, "ok")
	}
	return http.HandlerFunc(fn)
}

func TestRequestIDMiddleware(t *testing.T) {
	ts := httptest.NewServer(middleware.RequestIDMiddleware(testHandler()))
	defer ts.Close()

	client := ts.Client()

	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	assert.Nil(t, err)

	resp, err := client.Do(req)
	assert.Nil(t, err)

	reqId := resp.Header.Get(middleware.RequestIDHeader)
	assert.NotEqual(t, "", reqId)
}

func TestRequestIDMiddlewareWithValue(t *testing.T) {
	ts := httptest.NewServer(middleware.RequestIDMiddleware(testHandler()))
	defer ts.Close()

	client := ts.Client()

	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	assert.Nil(t, err)
	req.Header.Set(middleware.RequestIDHeader, "foobar")

	resp, err := client.Do(req)
	assert.Nil(t, err)

	reqId := resp.Header.Get(middleware.RequestIDHeader)
	assert.Equal(t, "foobar", reqId)
}
