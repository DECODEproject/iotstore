package tasks

import (
	"github.com/spf13/cobra"

	"github.com/thingful/iotstore/pkg/logger"
	"github.com/thingful/iotstore/pkg/server"
)

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringP("addr", "a", "0.0.0.0:8080", "The address to which the server binds")
	serverCmd.Flags().Bool("verbose", false, "Enable verbose output")
	serverCmd.Flags().StringP("cert-file", "c", "", "The path to a TLS certificate file to enable TLS on the server")
	serverCmd.Flags().StringP("key-file", "k", "", "The path to a TLS private key file to enable TLS on the server")
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
	},
}
