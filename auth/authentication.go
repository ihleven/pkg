package auth

import (
	"embed"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/ihleven/errors"
	"golang.org/x/crypto/bcrypt"
)

type Authentication interface {
	LoginHandler(w http.ResponseWriter, r *http.Request)
	LogoutHandler(w http.ResponseWriter, r *http.Request)
}

func NewAuthentication() Authentication {
	a := authentication{make(map[string]*Account)}
	a.addAccount(&Matthias)
	a.addAccount(&Wolfgang)
	return &a
}

type authentication struct {
	Users map[string]*Account
}

func (a *authentication) addAccount(account *Account) {
	a.Users[account.Username] = account
}

func (a *authentication) AuthenticateUser(username, password string) (*Account, error) {

	account, ok := a.Users[username]
	if ok && account != nil {
		bcrypterr := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
		// If a password exists for the given user
		// AND, if it is the same as the password we received, the we can move ahead
		// if NOT, then we return an "Unauthorized" status
		if bcrypterr == nil {
			return account, nil
		}

		// tmp workaround für plaintext passwords
		if account.Password == password {
			return account, nil
		}

	}
	return nil, errors.NewWithCode(http.StatusUnauthorized, "Invalid credentials")
}

func (a *authentication) Authenticate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	account, err := a.AuthenticateUser(r.PostFormValue("username"), r.PostFormValue("password"))
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}

	expirationTime := time.Now().Add(500 * time.Hour)
	token, err := TokenString(account.Username, expirationTime)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		http.Error(w, err.Error(), 500)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  expirationTime,
		Secure:   false, //scheme == "https",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	http.Redirect(w, r, "/home", 301)
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(account)

}

//go:embed templates/*
var templates embed.FS

func (a *authentication) LoginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		r.ParseForm()
		account, err := a.AuthenticateUser(r.PostFormValue("username"), r.PostFormValue("password"))
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		expirationTime := time.Now().Add(500 * time.Hour)
		token, err := TokenString(account.Username, expirationTime)
		if err != nil {
			// If there is an error in creating the JWT return an internal server error
			http.Error(w, err.Error(), 500)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			Expires:  expirationTime,
			Secure:   false, //scheme == "https",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
		})
		if accept := r.Header.Get("Accept"); strings.HasPrefix(accept, "application/json") {
			json.NewEncoder(w).Encode(struct {
				Token   string   `json:"token"`
				Account *Account `json:"account"`
			}{Token: token, Account: account})
		} else {
			http.Redirect(w, r, "/home", 301)
			// w.Header().Set("Content-Type", "application/json")
			// w.WriteHeader(http.StatusOK)
			// json.NewEncoder(w).Encode(account)
		}
		return
	}

	if accept := r.Header.Get("Accept"); !strings.HasPrefix(accept, "text/html") {
		w.WriteHeader(405) // MethodNotAllowed
	} else {

		t, err := template.ParseFS(templates, "templates/*.html")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(200)
		t.ExecuteTemplate(w, "login.html", nil)
	}

}
func (a *authentication) LogoutHandler(w http.ResponseWriter, r *http.Request) {

	c := &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		Secure:   true,
		HttpOnly: true,
	}

	http.SetCookie(w, c)
	http.Redirect(w, r, "/login", 301)
}
