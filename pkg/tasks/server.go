package tasks

import (
	"context"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/lestrrat-go/backoff"
	"github.com/spf13/cobra"

	"github.com/DECODEproject/iotstore/pkg/logger"
	"github.com/DECODEproject/iotstore/pkg/server"
	"github.com/DECODEproject/iotstore/pkg/version"
)

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringP("addr", "a", "0.0.0.0:8080", "The address to which the server binds")
	serverCmd.Flags().StringP("cert-file", "c", "", "The path to a TLS certificate file to enable TLS on the server")
	serverCmd.Flags().StringP("key-file", "k", "", "The path to a TLS private key file to enable TLS on the server")

	raven.SetRelease(version.Version)
	raven.SetTagsContext(map[string]string{"component": "datastore"})
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts datastore listening for requests",
	Long: `
Starts our implementation of the DECODE datastore RPC interface, which is
designed to expose a simple API to store and retrieve encrypted events coming
from upstream IoT devices.

The server uses Twirp to expose both a JSON API along with a more performant
Protocol Buffer API. The JSON API is not intended for use other than for
clients unable to use the Protocol Buffer API.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := cmd.Flags().GetString("addr")
		if err != nil {
			return err
		}

		datasource, err := GetFromEnv("IOTSTORE_DATABASE_URL")
		if err != nil {
			return err
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}

		certFile, err := cmd.Flags().GetString("cert-file")
		if err != nil {
			return err
		}

		keyFile, err := cmd.Flags().GetString("key-file")
		if err != nil {
			return err
		}

		logger := logger.NewLogger()

		e := backoff.ExecuteFunc(func(_ context.Context) error {
			s := server.NewServer(
				&server.Config{
					Addr:     addr,
					ConnStr:  datasource,
					Verbose:  verbose,
					CertFile: certFile,
					KeyFile:  keyFile,
				},
				logger,
			)

			return s.Start()
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		policy := backoff.NewExponential()
		return backoff.Retry(ctx, policy, e)
	},
}
