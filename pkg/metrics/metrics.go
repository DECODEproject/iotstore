package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// MustRegister is a wrapper around prometheus.MustRegister that attempts a
// basic retry. We need this for the retry loop in the main server entrypoint
// which can retry quickly if postgres is not available.
func MustRegister(c prometheus.Collector) {
	err := prometheus.Register(c)
	if err != nil {
		if prometheus.Unregister(c) {
			prometheus.MustRegister(c)
		} else {
			panic(err)
		}
	}
}
