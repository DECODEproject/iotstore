package tasks

import (
	"github.com/spf13/cobra"

	"github.com/thingful/iotstore/pkg/server"
)

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringP("addr", "a", "0.0.0.0:8080", "Specify the address to which the server binds")
	serverCmd.Flags().StringP("datasource", "d", "", "Database connection URL")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start server listening for requests",
	Long: `Starts a simple http server to verify that the process runs
continuously within the container.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := cmd.Flags().GetString("addr")
		if err != nil {
			return err
		}

		datasource, err := cmd.Flags().GetString("datasource")
		if err != nil {
			return err
		}

		s := server.NewServer(addr, datasource)

		s.Start()

		return nil
	},
}
