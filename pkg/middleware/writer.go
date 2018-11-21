package middleware

import (
	"net/http"
)

// writer is a struct that wraps http.ResponseWriter to allow capturing and
// exposing the status code.
type writer struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code, and then calls the wrapped response
// writer method.
func (w *writer) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
