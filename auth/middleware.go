package auth

import (
	"context"
	"net/http"

	"github.com/ihleven/errors"
)

// HTTP middleware setting a value on the request context
func Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == "/signin" {
			next.ServeHTTP(w, r)
			return
		}

		claims, status, err := GetClaims(r)
		// fmt.Printf("*** claims error %d: %v => %v\n", status, claims, err)
		if err != nil {

			// http.Redirect(w, r, "/login", 302)
			w.WriteHeader(status)
			// fmt.Fprintf(w, "*** claims error %d: %v => %v", status, claims, err)
			http.Error(w, err.Error(), errors.Code(err))

		} else {
			// fmt.Printf("*** claims error %d: %v => %v\n", status, claims, err)
			ctx := context.WithValue(r.Context(), "username", claims.Username)

			next.ServeHTTP(w, r.WithContext(ctx))

		}
	}
}
