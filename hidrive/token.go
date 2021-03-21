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
		authmap:     readAuthMapFromFile(), // make(map[string]*Auth),
		oauthClient: NewOAuth2Client(id, secret),
	}

	// authmap := mgmt.

	fmt.Println("NewAuthManager:", mgmt.authmap)

	return &mgmt
}

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
	fmt.Println("old:", auth.Token.AccessToken, auth.Token.RefreshToken, err)
	fmt.Println("new:", new.AccessToken, new.RefreshToken, err)

	auth.Token = new
	m.writeTokenFile()

	return auth.Token.AccessToken, nil
}

func (m *AuthManager) GetAuth(key string) *Auth {
	m.Lock()
	defer m.Unlock()
	return m.authmap[key]
}

type AuthToken struct {
	AccessToken string
	Alias       string
}

func (m *AuthManager) GetAuthToken(key string) (*AuthToken, error) {
	m.Lock()
	defer m.Unlock()
	if auth, ok := m.authmap[key]; ok {
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

func readAuthMapFromFile() map[string]*Auth {
	bytes, err := os.ReadFile("hidrive.mgmt")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return nil // errors.Wrap(err, "failed to read token file")
	}
	authmap := make(map[string]*Auth)
	if err := json.Unmarshal(bytes, &authmap); err != nil {
		return nil // errors.Wrap(err, "error parsing token")
	}
	fmt.Printf("readTokenFile() => %v\n", authmap)

	return authmap
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

		params := url.Values{
			"client_id":     {m.clientID},
			"response_type": {"code"},
			"scope":         {"user,rw"},
			"lang":          {"de"},                       // optional: language in which the authorization page is shown
			"state":         {r.URL.Query().Get("state")}, // optional:
			"redirect_uri":  {"http://localhost:8000/hidrive/auth/authcode?next=" + r.URL.Query().Get("next")},
		}
		http.Redirect(w, r, "https://my.hidrive.com/client/authorize?"+params.Encode(), 302)

	case "/authcode":
		// HandleAuthorizeCallback verarbeitet die Weiterleitung nach erfolgtem authorize
		// sollte unter  https://ihle.cloud/hidrive-token-callback erreichbar sein

		key := r.URL.Query().Get("state")
		if code, ok := r.URL.Query()["code"]; ok {

			_, err := m.AddAuth(key, code[0])
			if err != nil {
				fmt.Fprintf(w, "error => %+v\n", err)
			}
			// p.writeTokenFile()

			next := r.URL.Query().Get("next")
			if next == "" {
				next = "/hidrive/auth"
			}
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
