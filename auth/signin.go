package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func Handler() http.HandlerFunc {

	tokens := map[string]string{"rex3": "/hochzeit"}

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "POST" {
			r.ParseForm()
			token := r.PostFormValue("password")
			if _, ok := tokens[token]; !ok {
				w.WriteHeader(http.StatusUnauthorized)
			}
			expirationTime := time.Now().Add(100 * 24 * time.Hour)
			token, err := TokenString(token, expirationTime)
			if err != nil {
				// If there is an error in creating the JWT return an internal server error
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Finally, we set the client cookie for "token" as the JWT we just generated
			// we also set an expiry time which is the same as the token itself
			http.SetCookie(w, &http.Cookie{
				Name:     "token",
				Value:    token,
				Expires:  expirationTime,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
				Path:     "/",
			})

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(token))
		} else {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintln(w, "<form method='POST'><input name='token' /><button type='submit' /></form>")
			return
		}

	}
}

func SigninHandler(users map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("signin")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, token")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "GET" || r.Method == "OPTIONS" {
			// w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintln(w, "<form method='POST'><input name='username' /><input name='password' /><button type='submit' /></form>")
			return
		}

		// fmt.Println("origin:http://192.168.2.114:3000/", r.Header.Get("Origin"), r.Header.Get("X-Forwarded-Proto"))
		// w.Header().Set("Access-Control-Allow-Origin", "http://192.168.2.114:3000") //r.Header.Get("Origin"))
		// w.Header().Set("Access-Control-Allow-Credentials", "true")

		var credentials struct {
			Password string `json:"password"`
			Username string `json:"username"`
		}

		// https://medium.com/@edwardpie/processing-form-request-data-in-golang-2dff4c2441be
		contentType := r.Header.Get("Content-type")
		switch {
		// standard forms
		case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
			r.ParseForm()
			credentials.Username = r.PostFormValue("username")
			credentials.Password = r.PostFormValue("password")
		case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
			// username := r.FormValue("username")
			// password := r.FormValue("password")
		case strings.HasPrefix(contentType, "application/json"):
			err := json.NewDecoder(r.Body).Decode(&credentials)
			if err != nil {
				// If the structure of the body is wrong, return an HTTP error
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		fmt.Printf(" * parsed username %q and password %q", credentials.Username, credentials.Password)

		// Get the expected password from our in memory map
		hash, ok := users[credentials.Username]
		bcrypterr := bcrypt.CompareHashAndPassword([]byte(hash), []byte(credentials.Password))

		// If a password exists for the given user
		// AND, if it is the same as the password we received, the we can move ahead
		// if NOT, then we return an "Unauthorized" status
		if !ok || bcrypterr != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fmt.Println("signin success")
		expirationTime := time.Now().Add(500 * time.Hour)
		token, err := TokenString(credentials.Username, expirationTime)
		if err != nil {
			// If there is an error in creating the JWT return an internal server error
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Println("scheme: ", r.URL.Scheme, " - ", r.Header.Get("X-Forwarded-Proto"))
		// scheme := r.Header.Get("X-Forwarded-Proto")
		// Finally, we set the client cookie for "token" as the JWT we just generated
		// we also set an expiry time which is the same as the token itself
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			Expires:  expirationTime,
			Secure:   false, //scheme == "https",
			HttpOnly: true,
			// SameSite: http.SameSiteStrictMode,
			Path: "/",
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(credentials.Username))
	}
}
func Welcome(w http.ResponseWriter, r *http.Request) {

	claims, _, _ := GetClaims(r)
	json.NewEncoder(w).Encode(claims)
	// w.Write([]byte(fmt.Sprintf("Welcome %s!", claims.Username)))
}

func SignoutHandler(w http.ResponseWriter, r *http.Request) {

	c := &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		Secure:   true,
		HttpOnly: true,
	}

	http.SetCookie(w, c)
}
