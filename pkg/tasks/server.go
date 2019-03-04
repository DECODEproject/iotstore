package tasks

import (
	"context"
	"errors"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/lestrrat-go/backoff"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/DECODEproject/iotstore/pkg/logger"
	"github.com/DECODEproject/iotstore/pkg/server"
	"github.com/DECODEproject/iotstore/pkg/version"
)

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringP("addr", "a", ":8080", "The address to which the server binds")
	serverCmd.Flags().StringP("database-url", "d", "", "URL at which Postgres is listening (e.g. postgres://username:password@host:5432/dbname?sslmode=enable)")
	serverCmd.Flags().StringSlice("domains", []string{}, "Comma separated list of domains we will obtain TLS certificates for")
	serverCmd.Flags().Bool("verbose", false, "Enable verbose output")

	viper.BindPFlag("addr", serverCmd.Flags().Lookup("addr"))
	viper.BindPFlag("database-url", serverCmd.Flags().Lookup("database-url"))
	viper.BindPFlag("domains", serverCmd.Flags().Lookup("domains"))
	viper.BindPFlag("verbose", serverCmd.Flags().Lookup("verbose"))

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
clients unable to use the Protocol Buffer API.

Configuration values can be provided either by flags shown below, or via
environment variables. If a flag is named: --example-flag, then it also be
able to be supplied via an environment variable, i.e.: $IOTSTORE_EXAMPLE_FLAG

The server natively supports TLS via LetsEncrypt so if any domain names are
passed in via the domains flag, the server will attempt to obtain
certificates for the given domains, and start in TLS mode. If this list is
empty the server will start in non-TLS mode. Please note that the LetsEncrypt
provided certificate handshake will only work if the server is running, and
routable at the domains specified.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := viper.GetString("addr")
		if addr == "" {
			return errors.New("Must provide bind address")
		}

		databaseURL := viper.GetString("database-url")
		if databaseURL == "" {
			return errors.New("Must provide database url")
		}

		logger := logger.NewLogger()

		e := backoff.ExecuteFunc(func(_ context.Context) error {
			s := server.NewServer(
				&server.Config{
					Addr:    addr,
					ConnStr: databaseURL,
					Verbose: viper.GetBool("verbose"),
					Domains: viper.GetStringSlice("domains"),
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
