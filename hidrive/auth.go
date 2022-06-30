package hidrive

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ihleven/errors"
)

// NewTokenManager constructs an auth provider
func NewAuthManager(id, secret string) *AuthManager {

	mgmt := AuthManager{
		clientID:     id,
		clientSecret: secret,
		oauthClient:  NewOAuth2Client(id, secret),
	}
	var err error
	mgmt.authmap, err = readAuthMapFromFile()
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	return &mgmt
}

// AuthManager verwaltet auth-Daten für angemeldete user.
// Annahme: user-Anmeldungen werden außerhalb validiert. Übergebene authkeys sind authorisiert.
// Zur Kommunikation mit der HiDrive-APi wird ein OAuth2Client verwendet
type AuthManager struct {
	sync.Mutex
	authmap      map[string]*Auth `json:"-"`
	oauthClient  *OAuth2Client    `json:"-"`
	clientID     string           `json:"-"`
	clientSecret string           `json:"-"`
}

// Auth enthält alle Daten, die im AuthManager für einen key gespeichert werden.
type Auth struct {
	Token  *Token    `json:"token"`
	Expiry time.Time `json:"expiry,omitempty"`
}

func (m *AuthManager) GetAccessToken(key string) (*Token, error) {

	m.Lock()
	defer m.Unlock()
	auth, ok := m.authmap[key]
	if !ok || auth == nil {
		return nil, errors.NewWithCode(401, "Unknown auth key")
	}
	remaining := time.Until(auth.Expiry).Minutes()

	switch {
	case remaining < 1:
		// abgelaufen, blocking refresh
		new, err := m.oauthClient.RefreshToken(auth.Token.RefreshToken)
		if err != nil {
			return nil, errors.Wrap(err, "Error refresh")
		}
		auth.Token = new
		auth.Expiry = time.Now().Add(time.Second * time.Duration(new.ExpiresIn))
		m.writeTokenFile()
		fmt.Println(" -> blocking refresh", auth.Expiry)
	// case remaining < 5:
	// fast abgelaufen, refresh im Hintergrund und altes Token direkt zurückgeben

	default:
		// Token gültig und wirdzurückgeben
		// fmt.Println(" -> OK", time.Now(), remaining, auth.Expiry)
	}

	return auth.Token, nil
}

func (m *AuthManager) Refresh(key string, withoutLock bool) (*Token, error) {
	if !withoutLock {
		m.Lock()
		defer m.Unlock()
	}

	auth, ok := m.authmap[key]
	if !ok || auth == nil {
		return nil, errors.NewWithCode(401, "Unknown auth key")
	}
	new, err := m.oauthClient.RefreshToken(auth.Token.RefreshToken)
	if err != nil {
		return nil, errors.Wrap(err, "Error refresh")
	}
	fmt.Printf(" -> refreshing token %s => %s\n", auth.Token.AccessToken, new.AccessToken)
	// fmt.Println("new:", new.AccessToken, new.RefreshToken, err)

	auth.Token = new
	auth.Expiry = time.Now().Add(time.Second * time.Duration(new.ExpiresIn))
	m.authmap[key] = auth
	m.writeTokenFile()

	return auth.Token, nil
}

func (m *AuthManager) AddAuth(key, code string) (*Auth, error) {

	token, err := m.oauthClient.GenerateToken(code)
	if err != nil {
		return nil, errors.Wrap(err, "Error generating token")
	}

	auth := Auth{
		Token:  token,
		Expiry: time.Now().Add(time.Second * time.Duration(token.ExpiresIn)),
	}

	m.Lock()
	defer m.Unlock()
	m.authmap[key] = &auth
	m.writeTokenFile()
	return &auth, nil
}

func (m *AuthManager) DelAuth(key string) error {
	m.Lock()
	defer m.Unlock()
	delete(m.authmap, key)
	m.writeTokenFile()
	return nil
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
