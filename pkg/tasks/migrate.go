package tasks

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thingful/iotstore/pkg/postgres"
)

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateNewCmd)
	migrateCmd.AddCommand(migrateDownCmd)

	migrateCmd.PersistentFlags().String("dir", "pkg/migrations", "Migrations directory")
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage Postgres migrations",
	Long: `This task provides subcommands for working with migrations for Postgres.

Up migrations are run automatically when the application boots, but here we
also offer commands to create properly named migration files, and a command
to run down migrations.`,
}

var migrateNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new Postgres migration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := cmd.Flags().GetString("dir")
		if err != nil {
			return err
		}

		return postgres.NewMigration(dir, args[0])
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Run down migrations against Postgres",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Migrating down")

		return nil
	},
}
