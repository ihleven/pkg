package hidrive

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ihleven/errors"
)

type OAuth2Prov struct {
	// config    AppConfig    `json:"-"`
	client    *http.Client `json:"-"`
	tokenFile string       `json:"-"`
	// tokens    map[string]Token     `json:"-"`
	// deadlines map[string]time.Time `json:"-"`

	mu sync.Mutex // guards cache
	// cache  map[string]*entry
	tokens map[string]*tokenmgmt
}

type tokenmgmt struct {
	*Token
	sync.RWMutex
	err      error
	deadline time.Time
	ready    chan struct{} // closed when res is ready
}

func (p *OAuth2Prov) GetAccessToken(key string) (string, error) {
	p.mu.Lock()
	mgmt, ok := p.tokens[key]
	p.mu.Unlock()
	fmt.Printf("GetAccessToken => %+v\n", mgmt)

	if !ok {
		return "", errors.New("no mgmt entry for key %s", key)
	}

	if mgmt == nil {
		return "", errors.New("empty mgmt entry for key %s", key)
	}

	remaining := mgmt.deadline.Sub(time.Now()).Minutes()

	// fmt.Printf(" ***************** %v -> %v **************\n\n\n", time.Now(), mgmt.deadline)
	if remaining <= 0 {
		p.mu.Lock()
		// refresh and wait for result
		// newmgmt, err := p.RefreshMgmt(mgmt)
		// if err == nil {
		// 	p.tokens[key] = newmgmt
		// 	mgmt = newmgmt
		// 	remaining = mgmt.deadline.Sub(time.Now()).Minutes()
		// 	p.mu.Unlock()
		// 	return mgmt.AccessToken, nil
		// } else {
		// 	p.mu.Unlock()
		// 	return "", errors.Wrap(err, "unable to refresh")
		// }
	}

	if remaining > 0 {
		if remaining < 10 {
			// fmt.Printf(" ***************** triggering refesh: %s -> %f **************\n", key, remaining)
			// refresh token in background
			go func(mgmt *tokenmgmt) {
				// newmgmt, err := p.RefreshMgmt(mgmt)
				// if err == nil {
				// 	p.mu.Lock()
				// 	p.tokens[key] = newmgmt
				// 	mgmt = newmgmt
				// 	remaining = mgmt.deadline.Sub(time.Now()).Minutes()
				// 	p.mu.Unlock()
				// } else {
				// 	log.Fatal("failed token refesh")
				// }
			}(mgmt)
		}
		return mgmt.AccessToken, nil
	}

	return "", errors.New("Could not retrieve access token for key %s", key)
}

type Token struct {
	// AccessToken is the token that authorizes and authenticates the requests.
	AccessToken string `json:"access_token"`
	// TokenType is the type of token.
	// The Type method returns either this or "Bearer", the default.
	TokenType string `json:"token_type,omitempty"`
	// RefreshToken is a token that's used by the application (as opposed to the user) to refresh the access token if it expires.
	RefreshToken string `json:"refresh_token,omitempty"`

	ExpiresIn int    `json:"expires_in"`
	UserID    string `json:"userid"`
	Alias     string `json:"alias"`
	// Expiry is the optional expiration time of the access token.
	//
	// If zero, TokenSource implementations will reuse the same
	// token forever and RefreshToken or equivalent
	// mechanisms for that TokenSource will not be used.
	// Expiry time.Time `json:"expiry,omitempty"`
	Scope string
}
