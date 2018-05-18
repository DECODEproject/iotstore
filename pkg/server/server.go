package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	kitlog "github.com/go-kit/kit/log"
	datastore "github.com/thingful/twirp-datastore-go"

	"github.com/thingful/iotstore/pkg/twirp"
)

// Server is our top level type, contains all other components, is responsible
// for starting and stopping them in the correct order.
type Server struct {
	srv    *http.Server
	ds     *twirp.Datastore
	logger kitlog.Logger
}

// NewServer returns a new simple HTTP server.
func NewServer(addr string, connStr string, logger kitlog.Logger) *Server {
	ds := twirp.NewDatastore(connStr, logger)
	twirpHandler := datastore.NewDatastoreServer(ds, nil)

	// create our http.Server instance
	srv := &http.Server{
		Addr:    addr,
		Handler: twirpHandler,
	}

	// return the instantiated server
	return &Server{
		srv:    srv,
		ds:     ds,
		logger: kitlog.With(logger, "module", "server"),
	}
}

// Start starts the server running. We also create a channel listening for
// interrupt signals before gracefully shutting down.
func (s *Server) Start() error {
	err := s.ds.Start()
	if err != nil {
		return err
	}

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	go func() {
		s.logger.Log("addr", s.srv.Addr, "msg", "starting")
		if err := s.srv.ListenAndServe(); err != nil {
			s.logger.Log("err", err)
			os.Exit(1)
		}
	}()

	<-stopChan
	return s.Stop()
}

func (s *Server) Stop() error {
	s.logger.Log("msg", "stopping")
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	err := s.ds.Stop()
	if err != nil {
		return err
	}

	return s.srv.Shutdown(ctx)
}
