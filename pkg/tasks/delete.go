package tasks

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/thingful/iotstore/pkg/logger"
	"github.com/thingful/iotstore/pkg/postgres"
)

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringP("before", "b", "", "timestamp expressed as a RFC3339/ISO8601 string")
	deleteCmd.MarkFlagRequired("before")
	deleteCmd.Flags().BoolP("execute", "e", false, "boolean flag that if set executes the deletion")
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete old data from the store",
	Long: `This task allows the operator to delete old events from the datastore.

The datastore currently permanently stores all incoming events into a
PostgreSQL database, so this helper command is intended to allow the operator
of the system to remove data that has been previously stored in order to free
up space. It is the callers responsiblity to ensure data is adequately backed
up as this command will irrevocably delete records from PostgreSQL.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		connStr, err := GetFromEnv(ConnStrKey)
		if err != nil {
			return errors.Wrap(err, "failed to get environment variable")
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return errors.Wrap(err, "failed to read verbose flag")
		}

		execute, err := cmd.Flags().GetBool("execute")
		if err != nil {
			return errors.Wrap(err, "failed to read execute flag")
		}

		before, err := cmd.Flags().GetString("before")
		if err != nil {
			return errors.Wrap(err, "failed to read required \"before\" parameter")
		}

		beforeTime, err := time.Parse(time.RFC3339, before)
		if err != nil {
			return err
		}

		logger := logger.NewLogger()

		db := postgres.NewDB(connStr, verbose, logger)

		err = db.Start()
		if err != nil {
			return err
		}

		return db.DeleteData(beforeTime, execute)
	},
}
