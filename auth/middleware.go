package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ihleven/errors"
)

func Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Println(r.Cookie("token"))
		claims, status, err := GetClaims(r)
		fmt.Printf("*** claims error %d: %v => %v\n", status, claims, err)
		if err != nil {

			http.Redirect(w, r, "/login", 302)
			w.WriteHeader(status)
			fmt.Fprintf(w, "*** claims error %d: %v => %v", status, claims, err)
		} else {
			fmt.Printf("*** claims error %d: %v => %v\n", status, claims, err)
			ctx := context.WithValue(r.Context(), "props", claims)

			next.ServeHTTP(w, r.WithContext(ctx))

		}
	}
}

func AuthMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		claims, status, err := GetClaims(r)
		if err != nil {

			// http.Redirect(w, r, "/login", 302)
			w.WriteHeader(status)
			http.Error(w, err.Error(), errors.Code(err))
		} else {
			ctx := context.WithValue(r.Context(), "claims", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}
