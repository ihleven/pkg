package hidrive

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sync"

	"github.com/ihleven/errors"
)

type AuthI interface {
}

type Token struct {
	sync.RWMutex
	Client     *OAuth2Client
	oauthtoken *OAuth2Token
}

func (t *Token) AccessToken() string {

	return ""
}

// RefreshToken may be used to generate a valid access_token anytime, using an existing and valid refresh_token.
func (t *Token) Refresh(clientID, clientSecret string) error {

	params := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {t.oauthtoken.RefreshToken},
	}

	response, err := (&http.Client{}).PostForm("https://my.hidrive.com/oauth2/token", params)
	if err != nil {
		return errors.Wrap(err, "Failed to call /oauth2/token")
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return NewOAuth2Error(response)
	}

	var token OAuth2Token
	err = json.NewDecoder(response.Body).Decode(&token)
	if err != nil {
		return errors.Wrap(err, "failed to decode response")
	}
	t.oauthtoken = &token
	return nil
}
