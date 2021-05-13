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

// OAuth2Client is responsible for cummunicating with the hidrive oauth2 endpoints:
// GET /client/authorize
// POST /oauth2/token
// POST /oauth2/tokeninfo
// POST /oauth2/revoke
type OAuth2Client struct {
	http.Client
	baseURL                string
	clientID, clientSecret string
}

// OAuth2Token is the payload of a hidrive token returned by POST /oauth2/token request
// AccessToken is used for api requests
// RefreshToken is used to renew abgelaufene tokens
type OAuth2Token struct {
	// TokenType is the type of token.
	// The Type method returns either this or "Bearer", the default.
	TokenType string `json:"token_type,omitempty"`
	// AccessToken is the token that authorizes and authenticates the requests.
	AccessToken string `json:"access_token"`
	// RefreshToken is a token that's used by the application (as opposed to the user) to refresh the access token if it expires.
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	UserID       string `json:"userid"`
	Alias        string `json:"alias"`
	Scope        string `json:"scope"`
}

// GenerateToken allows you to retrieve a new refresh_token following initial “code” flow authorization.
// wraps hidrive POST /oauth2/token request.
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

// RefreshToken may be used to generate a valid access_token anytime, using an existing and valid refresh_token.
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

// RevokeToken lets you revoke an active access_token or refresh_token.
// Revoking a refresh_token will also invalidate all related access_token.
// Revoking the refresh_token is the easiest way to accomplish a logout mechanism for your app.
// This is a functionality we explicitly encourage you to implement.
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

// TokenInfo calls an endpoint you may use to get information about your current access_token.
// The response will include the granted scope, expiry time, user alias and your client_id.
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
