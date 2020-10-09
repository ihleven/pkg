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
