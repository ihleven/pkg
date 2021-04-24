package auth

import (
	"context"
	"fmt"
	"net/http"
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
