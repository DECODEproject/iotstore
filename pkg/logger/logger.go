package logger

import (
	"os"

	"github.com/DECODEproject/iotstore/pkg/version"
	kitlog "github.com/go-kit/kit/log"
)

// NewLogger is a simple helper function that returns a kitlog.Logger instance
// ready for use.
func NewLogger() kitlog.Logger {
	logger := kitlog.NewJSONLogger(kitlog.NewSyncWriter(os.Stdout))
	return kitlog.With(logger, "service", version.BinaryName, "ts", kitlog.DefaultTimestampUTC)
}
