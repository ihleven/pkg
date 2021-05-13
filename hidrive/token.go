package hidrive

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/ihleven/errors"
)

// file in auth.go umbenennen

type Auth struct {
	UserID        string           `json:"userid"`
	Alias         string           `json:"alias"`
	Scope         string           `json:"scope"`
	Token         *OAuth2Token     `json:"token"`
	Info          *OAuth2TokenInfo `json:"info"`
	ExpiresAt     time.Time        `json:"expiry,omitempty"`
	RefreshExpiry time.Time        `json:"-"`
}

// NewTokenManager constructs an auth provider
func NewAuthManager(id, secret string) *AuthManager {

	mgmt := AuthManager{
		clientID:     id,
		clientSecret: secret,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		// tokenFile: "hidrive.tokens",
		// authmap:     readAuthMapFromFile(), // make(map[string]*Auth),
		oauthClient: NewOAuth2Client(id, secret),
	}
	var err error
	mgmt.authmap, err = readAuthMapFromFile()
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	// authmap := mgmt.

	fmt.Println("NewAuthManager:", mgmt.authmap)

	return &mgmt
}

// AuthManager verwaltet auth-Daten für angemeldete user.
// Annahme: user-Anmeldungen werden außerhalb validiert. Übergebene authkeys sind authorisiert.
// Zur Kommunikation mit der HiDrive-APi wird ein OAuth2Client verwendet
type AuthManager struct {
	sync.Mutex
	authmap      map[string]*Auth `json:"-"`
	client       *http.Client     `json:"-"`
	oauthClient  *OAuth2Client    `json:"-"`
	clientID     string           `json:"-"`
	clientSecret string           `json:"-"`
	// tokenFile    string             `json:"-"`
	// deadlines map[string]time.Time `json:"-"`
}

func (m *AuthManager) GetClientAuthorizeURL(state, next string) string {

	redirectURI := "http://localhost:8000/hidrive/auth/authcode"
	if next != "" {
		redirectURI += "?next=" + next
	}

	params := url.Values{
		"client_id":     {m.clientID},
		"response_type": {"code"},
		"scope":         {"user,rw"},
		"lang":          {"de"},  // optional: language in which the authorization page is shown
		"state":         {state}, // optional:
		"redirect_uri":  {redirectURI},
	}

	return "https://my.hidrive.com/client/authorize?" + params.Encode()
}

func (m *AuthManager) GetAuthTokenAlt(key string) (*AuthToken, error) {
	m.Lock()
	defer m.Unlock()
	if auth, ok := m.authmap[key]; ok {
		return &AuthToken{auth.Token.AccessToken, auth.Alias}, nil
	}
	return nil, errors.NewWithCode(401, "no token")
}

func (m *AuthManager) GetAccessToken(key string) (string, error) {
	m.Lock()
	defer m.Unlock()
	auth, ok := m.authmap[key]
	if !ok || auth == nil {
		return "", errors.NewWithCode(401, "Unknown auth key")
	}
	return auth.Token.AccessToken, nil
}

func (m *AuthManager) Refresh(key string) (string, error) {

	m.Lock()
	defer m.Unlock()

	auth, ok := m.authmap[key]
	if !ok || auth == nil {
		return "", errors.NewWithCode(401, "Unknown auth key")
	}
	new, err := m.oauthClient.RefreshToken(auth.Token.RefreshToken)
	if err != nil {
		return "", errors.Wrap(err, "Error refresh")
	}
	fmt.Printf(" -> refreshing token %s => %s\n", auth.Token.AccessToken, new.AccessToken)
	// fmt.Println("new:", new.AccessToken, new.RefreshToken, err)

	auth.Token = new
	auth.ExpiresAt = time.Now().Add(time.Second * time.Duration(new.ExpiresIn))
	m.authmap[key] = auth
	m.writeTokenFile()

	return auth.Token.AccessToken, nil
}

// func (m *AuthManager) GetAuth(key string) *Auth {
// 	m.Lock()
// 	defer m.Unlock()
// 	return m.authmap[key]
// }

type AuthToken struct {
	AccessToken string
	Alias       string
}

func (m *AuthManager) GetAuthToken(key string) (*AuthToken, error) { //<-chan string {
	m.Lock()
	defer m.Unlock()
	if auth, ok := m.authmap[key]; ok {

		remaining := auth.ExpiresAt.Sub(time.Now()).Minutes()
		if remaining < 1 {
			fmt.Println("refreshing access token!!!")
			// access, err := m.Refresh(key)
			new, err := m.oauthClient.RefreshToken(auth.Token.RefreshToken)
			if err != nil {
				return nil, errors.NewWithCode(401, "no token")
			}
			auth.Token = new
			auth.ExpiresAt = time.Now().Add(time.Second * time.Duration(new.ExpiresIn))
			m.authmap[key] = auth
			m.writeTokenFile()
		}

		return &AuthToken{auth.Token.AccessToken, auth.Alias}, nil
	}

	return nil, errors.NewWithCode(401, "no token")
}

func (m *AuthManager) AddAuth(key, code string) (*Auth, error) {

	token, err := m.oauthClient.GenerateToken(code)
	if err != nil {
		return nil, errors.Wrap(err, "Error generating token")
	}

	auth := Auth{
		UserID:    token.UserID,
		Alias:     token.Alias,
		Scope:     token.Scope,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Second * time.Duration(token.ExpiresIn)),
	}

	m.Lock()
	defer m.Unlock()
	m.authmap[key] = &auth
	m.writeTokenFile()
	return &auth, nil
}

func readAuthMapFromFile() (map[string]*Auth, error) {

	authmap := make(map[string]*Auth)

	bytes, err := os.ReadFile("hidrive.mgmt")
	if err != nil {
		if os.IsNotExist(err) {
			return authmap, nil //errors.Wrap(err, "")
		}
		return nil, errors.Wrap(err, "failed to read token file")
	}

	if err := json.Unmarshal(bytes, &authmap); err != nil {
		return nil, errors.Wrap(err, "error parsing token")
	}
	fmt.Printf("readTokenFile() => %v\n", authmap)

	return authmap, nil
}

func (m *AuthManager) writeTokenFile() error {

	bytes, err := json.MarshalIndent(m.authmap, "", "    ")
	if err != nil {
		return errors.Wrap(err, "error marshaling authmap")
	}
	err = os.WriteFile("hidrive.mgmt", bytes, 0664)
	if err != nil {
		return err
	}
	return nil
}

//go:embed templates/*
var templates embed.FS

// ServeHTTP has 3 modes:
// 1) GET without params shows a button calling ServeHTTP with the authorize param
//    and a form POSTing ServeHTTP with a code
// 2) GET with authorize param redirects to the hidrive /client/authorize endpoint where a user can authorize the app.
//    This endpoint will call registered token-callback which is not reachable locally ( => copy code and use form from 1)
// 3) POSTing username and code triggering oauth2/token endpoint generating an access token
func (m *AuthManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.URL.Path {

	case "/authorize":
		clientAuthURL := m.GetClientAuthorizeURL(r.URL.Query().Get("state"), r.URL.Query().Get("next"))

		http.Redirect(w, r, clientAuthURL, 302)

	case "/authcode":
		// HandleAuthorizeCallback verarbeitet die Weiterleitung nach erfolgtem authorize
		// sollte unter  https://ihle.cloud/hidrive-token-callback erreichbar sein

		key := r.URL.Query().Get("state")
		if code, ok := r.URL.Query()["code"]; ok {
			fmt.Println("code", code, ok)
			_, err := m.AddAuth(key, code[0])
			if err != nil {
				fmt.Printf("error => %+v\n", err)
				fmt.Fprintf(w, "error => %+v\n", err)
			}
			// p.writeTokenFile()

			next := r.URL.Query().Get("next")
			if next == "" {
				next = "/hidrive/auth"
			}
			fmt.Println("code", code, ok, next)
			fmt.Fprintf(w, `<head><meta http-equiv="Refresh" content="0; URL=%s"></head>`, next)
		}

	default:

		if key, found := r.URL.Query()["refresh"]; found {
			m.Refresh(key[0])
		}

		t, err := template.ParseFS(templates, "templates/*.html")
		if err != nil {
			fmt.Fprintf(w, "Cannot parse templates: %v", err)
			return
		}

		w.WriteHeader(200)
		t.ExecuteTemplate(w, "tokenmgmt.html", map[string]interface{}{"ClientID": m.clientID, "ClientSecret": m.clientSecret, "tokens": m.authmap})
		return
	}
}
