package httpauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SigninHandler:
// 1. parse credentials
// 2. authenticate user with credentials
// 3. login user (e.g. via cookie/header)
func (a *Auth) SigninHandler(w http.ResponseWriter, r *http.Request) {
	// https://medium.com/@edwardpie/processing-form-request-data-in-golang-2dff4c2441be
	contentType := r.Header.Get("Content-type")

	fmt.Println("signin", contentType)

	var credentials struct {
		Password string `json:"password"`
		Username string `json:"username"`
	}

	switch {
	// standard forms
	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		fallthrough
	case strings.HasPrefix(contentType, "multipart/form-data"):
		r.ParseForm()
		credentials.Username = r.PostFormValue("username")
		credentials.Password = r.PostFormValue("password")
	// case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
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
	fmt.Printf(" * parsed username %q and password %q\n", credentials.Username, credentials.Password)

	account := a.Authenticate(credentials.Username, credentials.Password)
	if account == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	fmt.Printf(" * account -> %v\n", account)

	token := a.Login(w, account)

	// http.Redirect(w, r, "/home", 301)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	type foo struct {
		Username string
		JWT      string
	}
	bytes, _ := json.MarshalIndent(foo{credentials.Username, token}, "", "    ")
	w.Write(bytes)
}

func (a *Auth) WelcomeHandler(w http.ResponseWriter, r *http.Request) {

	authkey := a.ParseRequestAuth(r)
	w.Write([]byte(fmt.Sprintf("Welcome %s!", authkey)))
}

func (a *Auth) SignoutHandler(w http.ResponseWriter, r *http.Request) {

	c := &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		Secure:   false,
		HttpOnly: true,
	}
	fmt.Println("SignoutHandler")
	http.SetCookie(w, c)
	// http.Redirect(w, r, "/login", 301)
}
