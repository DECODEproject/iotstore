package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	datastore "github.com/thingful/twirp-datastore-go"

	"github.com/thingful/iotstore/pkg/twirp"
)

// Server is our top level type, contains all other components, is responsible
// for starting and stopping them in the correct order.
type Server struct {
	srv *http.Server
	ds  *twirp.Datastore
}

// NewServer returns a new simple HTTP server.
func NewServer(addr string, connStr string) *Server {
	ds := twirp.NewDatastore(connStr)
	twirpHandler := datastore.NewDatastoreServer(ds, nil)

	// create our http.Server instance
	srv := &http.Server{
		Addr:    addr,
		Handler: twirpHandler,
	}

	// return the instantiated server
	return &Server{
		srv: srv,
		ds:  ds,
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
		log.Println("Starting server")
		if err := s.srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	<-stopChan
	return s.Stop()
}

func (s *Server) Stop() error {
	log.Println("Stopping server")
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	err := s.ds.Stop()
	if err != nil {
		return err
	}

	return s.srv.Shutdown(ctx)
}
