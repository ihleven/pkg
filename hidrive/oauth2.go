package hidrive

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ihleven/errors"
)

// NewClient creates a new hidrive oauth2 client
func NewOAuth2Client(clientID, clientSecret string) *OAuth2Client {

	var OAuth2Client = OAuth2Client{
		Client: http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL:      "https://my.hidrive.com",
		clientID:     clientID,
		clientSecret: clientSecret,
	}
	return &OAuth2Client
}

type OAuth2Client struct {
	http.Client
	baseURL                string
	clientID, clientSecret string
}

type OAuth2Token struct {
	TokenType    string `json:"token_type,omitempty"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	UserID       string `json:"userid"`
	Alias        string `json:"alias"`
	Scope        string `json:"scope"`
}

func (c *OAuth2Client) GenerateToken(code string) (*OAuth2Token, error) {

	params := url.Values{
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
		"grant_type":    {"authorization_code"},
		"code":          {code},
	}

	response, err := c.PostForm(c.baseURL+"/oauth2/token", params)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to call /oauth2/token")
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return nil, NewOAuth2Error(response)
	}

	var token OAuth2Token
	err = json.NewDecoder(response.Body).Decode(&token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}
	return &token, nil
}

func (c *OAuth2Client) RefreshToken(refreshtoken string) (*OAuth2Token, error) {

	params := url.Values{
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshtoken},
	}

	response, err := c.PostForm(c.baseURL+"/oauth2/token", params)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to call /oauth2/token")
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return nil, NewOAuth2Error(response)
	}

	var token OAuth2Token
	err = json.NewDecoder(response.Body).Decode(&token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}
	fmt.Println(token)
	return &token, nil
}

func (c *OAuth2Client) RevokeToken(token string) error {

	params := url.Values{
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
		"token":         {token},
	}

	response, err := c.PostForm(c.baseURL+"/oauth2/revoke", params)
	if err != nil {
		return errors.Wrap(err, "Failed to call /oauth2/revoke")
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return NewOAuth2Error(response)
	}

	return nil
}

type OAuth2TokenInfo struct {
	ExpiresIn int    `json:"expires_in"`
	ClientID  string `json:"client_id"`
	Alias     string `json:"alias"`
	Scope     string `json:"scope"`
}

func (c *OAuth2Client) TokenInfo(accessToken string) (*OAuth2TokenInfo, error) {

	params := url.Values{
		"access_token": {accessToken},
	}

	response, err := c.PostForm(c.baseURL+"/oauth2/tokeninfo", params)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to call /oauth2/tokeninfo")
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return nil, NewOAuth2Error(response)
	}

	var info OAuth2TokenInfo
	err = json.NewDecoder(response.Body).Decode(&info)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode token info response")
	}
	return &info, nil
}
