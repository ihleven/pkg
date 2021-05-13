package hidrive

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/ihleven/errors"
)

// Auth enthält alle Daten, die im AuthManager für einen key gespeichert werden.
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
	fmt.Println("GAETASEFASDFASDFASDFSDF")
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
