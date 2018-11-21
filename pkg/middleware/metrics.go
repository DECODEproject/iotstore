package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "decode",
			Subsystem: "datastore",
			Name:      "request_duration_sec",
			Help:      "Time (in seconds) spent serving HTTP requests",
			Buckets:   prometheus.DefBuckets,
		}, []string{"status_code", "method", "path"},
	)
)

func init() {
	prometheus.MustRegister(requestDuration)
}

// MetricsMiddleware is a net.http compatible middleware that captures some metrics about every
// request which we then expose via prometheus.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// initialize as OK, other stages may change this
		cw := &writer{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		startTime := time.Now()
		next.ServeHTTP(cw, r)
		took := time.Since(startTime)
		requestDuration.WithLabelValues(
			strconv.Itoa(cw.statusCode), r.Method, r.URL.Path).
			Observe(took.Seconds())
	})
}
