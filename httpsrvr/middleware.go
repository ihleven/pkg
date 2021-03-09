package httpsrvr

import (
	"net/http"

	"golang.org/x/time/rate"
)

func limit(next http.Handler, limiter *rate.Limiter) http.Handler {
	if limiter != nil {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limiter.Allow() == false {
				http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	} else {
		return next
	}
}

// ResponseWriter intercepts http.ResponseWriter  capturing the response status code
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewResponseWriter wraps given http.ResponseWriter in a ResponseWriter overwriting the WriteHeader method
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	// WriteHeader(int) is not called if our response implicitly returns 200 OK, so
	// we default to that status code.
	return &ResponseWriter{w, http.StatusOK}
}

// WriteHeader captures the response status
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
