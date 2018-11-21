package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	kitlog "github.com/go-kit/kit/log"
	twrpprom "github.com/joneskoo/twirp-serverhook-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	datastore "github.com/thingful/twirp-datastore-go"
	goji "goji.io"
	pat "goji.io/pat"

	"github.com/DECODEproject/iotstore/pkg/middleware"
	"github.com/DECODEproject/iotstore/pkg/postgres"
	"github.com/DECODEproject/iotstore/pkg/rpc"
)

// Config is a struct used to pass in configuration from the calling task
type Config struct {
	Addr     string
	ConnStr  string
	Verbose  bool
	CertFile string
	KeyFile  string
}

// Server is our top level type, contains all other components, is responsible
// for starting and stopping them in the correct order.
type Server struct {
	srv    *http.Server
	ds     *rpc.Datastore
	logger kitlog.Logger
	config *Config
}

// PulseHandler is a function that closes over our DB instance returning an
// http.Handler that attempts to connect to the DB, returning either 200 OKk
// with an ok response if successful, or an error message with a 500 status if
// the DB connection failed.
func PulseHandler(db *postgres.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := db.Ping()
		if err != nil {
			http.Error(w, "failed to connect to DB", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "ok")
	})
}

// NewServer returns a new simple HTTP server.
func NewServer(config *Config, logger kitlog.Logger) *Server {
	ds := rpc.NewDatastore(config.ConnStr, config.Verbose, logger)
	hooks := twrpprom.NewServerHooks(nil)

	twirpHandler := datastore.NewDatastoreServer(ds, hooks)

	mux := goji.NewMux()

	// set up the handlers
	mux.Handle(pat.Post(datastore.DatastorePathPrefix+"*"), twirpHandler)
	mux.Handle(pat.Get("/pulse"), PulseHandler(ds.DB))
	mux.Handle(pat.Get("/metrics"), promhttp.Handler())

	// add our middleware
	mux.Use(middleware.RequestIDMiddleware)
	mux.Use(middleware.MetricsMiddleware)

	// create our http.Server instance
	srv := &http.Server{
		Addr:    config.Addr,
		Handler: mux,
	}

	// return the instantiated server
	return &Server{
		srv:    srv,
		ds:     ds,
		logger: kitlog.With(logger, "module", "server"),
		config: config,
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
		s.logger.Log(
			"listenAddr",
			s.srv.Addr,
			"msg",
			"starting server",
			"pathPrefix",
			datastore.DatastorePathPrefix,
			"tlsEnabled",
			isTLSEnabled(s.config),
		)

		if isTLSEnabled(s.config) {
			if err := s.srv.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile); err != nil {
				s.logger.Log("err", err)
				os.Exit(1)
			}
		} else {
			if err := s.srv.ListenAndServe(); err != nil {
				s.logger.Log("err", err)
				os.Exit(1)
			}
		}
	}()

	<-stopChan
	return s.Stop()
}

// Stop stops the server running.
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

// isTLSEnabled returns true if the passed in configuration object contains both
// a cert and key file paths, false otherwise.
func isTLSEnabled(config *Config) bool {
	return config.CertFile != "" && config.KeyFile != ""
}
