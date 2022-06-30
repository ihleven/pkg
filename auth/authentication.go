package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ihleven/errors"
	"golang.org/x/crypto/bcrypt"
)

type AuthenticationDep interface {
	LoginHandler(w http.ResponseWriter, r *http.Request)
	LogoutHandler(w http.ResponseWriter, r *http.Request)
}

func NewAuthenticator(options ...Option) AuthenticationDep {
	a := authentication{
		Users:  make(map[string]*Account),
		Tokens: make(map[string]interface{}),
	}

	for _, option := range options {
		option(&a)
	}

	return &a
}

type authentication struct {
	Users  map[string]*Account
	Tokens map[string]interface{}
}

type Option func(*authentication)

func User(account *Account) Option {
	return func(a *authentication) {
		a.Add(account)
	}
}

func (a *authentication) Add(account *Account) {
	a.Users[account.Username] = account
}

// AuthenticateUser given username is identical to stored password of Account
func (a *authentication) AuthenticateUser(username, password string) (*Account, error) {

	if account, ok := a.Users[username]; ok && account != nil {
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

// AuthenticateUser given username is identical to stored password of Account
func (a *authentication) AuthenticateToken(code string) (interface{}, error) {

	if token, ok := a.Tokens[code]; ok {

		return token, nil
	}

	return nil, errors.NewWithCode(http.StatusUnauthorized, "Invalid code")
}

func (a *authentication) LoginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {

		var credentials struct {
			Password string `json:"password"`
			Username string `json:"username"`
		}

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
		default:
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		fmt.Println("LoginHandler", credentials, a.Users[credentials.Username])
		account, err := a.AuthenticateUser(credentials.Username, credentials.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
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
	if accept := r.Header.Get("Accept"); !strings.HasPrefix(accept, "application/json") {
		// im nicht-JSON-Fall
		http.Redirect(w, r, "/login", 301)
	}

}
