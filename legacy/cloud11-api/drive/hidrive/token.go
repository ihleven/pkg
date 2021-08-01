package hidrive

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
}

func mustReadConfig() *Config {
	bytecontent, err := ioutil.ReadFile("./hidrive.config")
	if err != nil {
		panic(err)
	}
	var config Config
	if err := json.Unmarshal(bytecontent, &config); err != nil {
		panic(err)
	}
	return &config
}

func readToken() *Token {
	bytecontent, err := ioutil.ReadFile("./token")
	if err != nil {
		fmt.Println("error reading token file:", err)
		return nil
	}
	var token Token
	if err := json.Unmarshal(bytecontent, &token); err != nil {
		fmt.Println("error parsing token:", err)
		return nil
	}
	return &token
}
func writeToken(body []byte) error {

	err := ioutil.WriteFile("./token", body, 0644)
	if err != nil {
		return err
	}
	return nil
}

func NewToken() (*Token, error) {

	var config = mustReadConfig()

	token := readToken()

	if token != nil {

		fmt.Printf(" * token => %v\n", token.AccessToken)
		formData := url.Values{
			"client_id":     {config.ClientID},
			"client_secret": {config.ClientSecret},
			"grant_type":    {"refresh_token"},
			"refresh_token": {token.RefreshToken},
		}

		res, err := http.PostForm("https://my.hidrive.com/oauth2/token", formData)
		if err != nil {
			return nil, err
		}

		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		fmt.Printf(" * token refresh => %v\n", string(body))
		if err = json.Unmarshal(body, &token); err != nil {
			return nil, err
		}
		writeToken(body)
		return token, nil
	}

	formData := url.Values{
		"client_id":     {config.ClientID},
		"client_secret": {config.ClientSecret},
		"grant_type":    {"authorization_code"},
		"code":          {config.Code},
	}

	res, err := http.PostForm("https://my.hidrive.com/oauth2/token", formData)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	fmt.Printf(" * body => %v\n", string(body))

	if err = json.Unmarshal(body, &token); err != nil {
		return nil, err
	}
	fmt.Printf("NewToken() => %v\n", token)
	writeToken(body)
	return token, nil
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
	Expiry time.Time `json:"expiry,omitempty"`
}

func (t *Token) GetAccessToken() string {
	return t.AccessToken
}
