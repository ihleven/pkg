package httpsrvr

import (
	"net/http"
	"sync/atomic"
)

// ResponseWriter intercepts http.ResponseWriter  capturing the response status code
type ResponseWriter struct {
	http.ResponseWriter
	count      uint64
	statusCode int
	// status int
	// length int
}

// NewResponseWriter wraps given http.ResponseWriter in a ResponseWriter overwriting the WriteHeader method
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	// WriteHeader(int) is not called if our response implicitly returns 200 OK, so
	// we default to that status code.
	return &ResponseWriter{w, 0, http.StatusOK}
}

// Write captures the response size
func (rw *ResponseWriter) Write(buf []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(buf)
	atomic.AddUint64(&rw.count, uint64(n))
	return n, err
}

// WriteHeader captures the response status
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Count function return counted bytes
func (rw *ResponseWriter) Count() uint64 {
	return atomic.LoadUint64(&rw.count)
}
