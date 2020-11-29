package hidrive

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/ihleven/errors"
)

func NewOauthProvider() *OAuth2Prov {
	bytes, err := ioutil.ReadFile("hidrive.config")
	if err != nil {
		return nil
	}
	var config map[string]string
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		fmt.Println("err", err)
		return nil
	}
	// fmt.Println("bytes", string(bytes), config)

	oap := OAuth2Prov{
		ClientID:     config["client_id"],
		ClientSecret: config["client_secret"],
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		tokenFile: "hidrive.token",
	}
	err = oap.readTokenFile()
	if err != nil {
		fmt.Println("err", err)
		return nil
	}
	return &oap
}

type OAuth2Prov struct {
	ClientID      string `json:"client_id"`
	ClientSecret  string `json:"client_secret"`
	Token         Token
	client        *http.Client
	tokenFile     string
	TokenDeadline time.Time
}

func (p *OAuth2Prov) readTokenFile() error {
	bytes, err := ioutil.ReadFile(p.tokenFile)
	if err != nil {
		return errors.Wrap(err, "failed to read token file")
	}

	if err := json.Unmarshal(bytes, &p.Token); err != nil {
		return errors.Wrap(err, "error parsing token")
	}
	return nil
}

func (p *OAuth2Prov) writeTokenFile() error {
	bytes, err := json.Marshal(p.Token)
	if err != nil {
		return errors.Wrap(err, "error parsing token")
	}
	err = ioutil.WriteFile(p.tokenFile, bytes, 0664)
	if err != nil {
		return err
	}
	return nil
}

func (p *OAuth2Prov) GetAccessToken() string {

	now, deadline := time.Now(), p.TokenDeadline

	fmt.Printf("GetAccessToken (%v) => %v\n", deadline, now.After(deadline))
	if now.After(deadline) {
		fmt.Printf("%v: TokenDeadline expired (%v) => refreshing token\n", now, deadline)
		token, err := p.RefreshToken()
		if err != nil {
			return fmt.Sprintf("TokenDeadline expired, but error in RefreshToken %v", err)
		}
		p.TokenDeadline = now.Add(time.Second * time.Duration(token.ExpiresIn-200))
		fmt.Printf("new TokenDeadline => %v\n", p.TokenDeadline)
	}
	// if p.Token == nil {
	// 	return ""
	// }
	return p.Token.AccessToken
}

func (p *OAuth2Prov) TokenWithCode(code string) (*Token, error) {

	data := url.Values{
		"client_id":     {p.ClientID},
		"client_secret": {p.ClientSecret},
		"grant_type":    {"authorization_code"},
		"code":          {code},
	}

	resp, err := p.client.PostForm("https://my.hidrive.com/oauth2/token", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// bytes, err := ioutil.ReadAll(resp.Body)
	// fmt.Printf("bytes: %s\n", bytes)

	var token Token
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}
	p.Token = token
	return &token, nil
}

func (p *OAuth2Prov) RefreshToken() (*Token, error) {

	data := url.Values{
		"client_id":     {p.ClientID},
		"client_secret": {p.ClientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {p.Token.RefreshToken},
	}
	resp, err := p.client.PostForm("https://my.hidrive.com/oauth2/token", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var token Token
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}
	p.Token = token

	return &token, nil
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

type TokenInfo struct {
	ExpiresIn int    `json:"expires_in"`
	ClientID  string `json:"client_id"`
	Alias     string `json:"alias"`
	Scope     string `json:"scope"`
}

func (p *OAuth2Prov) TokenInfo() (*TokenInfo, error) {

	formData := url.Values{
		"access_token": {p.Token.AccessToken},
	}

	res, err := p.client.PostForm("https://my.hidrive.com/oauth2/tokeninfo", formData)
	if err != nil {
		return nil, errors.Wrap(err, "Error in tokeninfo request")
	}
	defer res.Body.Close()

	var info TokenInfo
	err = json.NewDecoder(res.Body).Decode(&info)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t parse tokeninfo response")
	}
	// fmt.Println("-------------------info", info, res.Status)
	return &info, nil
}

func (p OAuth2Prov) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// fmt.Fprintf(w, "hello: %v", r.URL)
	switch {
	case r.URL.Path == "/login":
		http.Redirect(w, r, "https://my.hidrive.com/client/authorize?client_id=b4436f1157043c2bf8db540c9375d4ed&response_type=code&scope=admin,rw", 302)
	case r.URL.Path == "/token":
		p.TokenPage(w, r)
	}
}

// authorize macht so keinen Sinn, sollte ein redirect sein.
func (p *OAuth2Prov) authorize() {

	req, err := http.NewRequest("GET", "https://my.hidrive.com/client/authorize", nil)
	if err != nil {
		os.Exit(1)
	}
	q := req.URL.Query()
	q.Add("client_id", p.ClientID)
	q.Add("response_type", "code")
	req.URL.RawQuery = q.Encode()

	// req.Header.Add("If-None-Match", `W/"wyzzy"`)
	resp, err := p.client.Do(req)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	fmt.Println(req.URL, string(bytes))
}

// Diese Methode verarbeitet die Weiterleitung nach erfolgtem authorize
func (p *OAuth2Prov) handleAuthorizeCallback(w http.ResponseWriter, r *http.Request) {
	// e.g. ?scope=rw,user&code=R8Pt3WTfbREUbx6SHnPn

	if scope, ok := r.URL.Query()["scope"]; ok {
		fmt.Fprintf(w, "scope => %q\n", scope)
	}
	code, ok := r.URL.Query()["code"]
	if ok {
		fmt.Fprintf(w, "code => %q\n", code)
		// err := p.tokenWithCode(code[0])
		token, err := p.TokenWithCode(code[0])
		if err == nil {
			fmt.Fprintf(w, "token => %+v\n", token)
		}
		p.writeTokenFile()
	}
}

// Token zeigt ein Eingabefeld für einen Authorize-Code und fordert damit ein Token an
func (p *OAuth2Prov) TokenPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		token, err := p.TokenWithCode(r.FormValue("code"))
		if err == nil {
			fmt.Fprintf(w, "token => %+v\n", token)
		}
		p.writeTokenFile()
	}
	tpl1 := `<html><body><form method="post"><input name="code"/></form></body></html>`
	fmt.Fprintf(w, tpl1)

}

// func (p *OAuth2Prov) GetToken(code string) error {

// 	var token *Token
// 	var err error
// 	if code != "" {

// 		token, err = p.authClient.PostOauth2Token(p.ClientID, p.ClientSecret, "authorization_code", code, "")
// 		if err != nil {
// 			return err
// 		}
// 	} else if p.Token != nil {
// 		token, err = p.authClient.PostOauth2Token(p.ClientID, p.ClientSecret, "refresh_token", "", p.Token.RefreshToken)
// 		if err != nil {
// 			return err
// 		}

// 	} else {
// 		return errors.New("no code or refresh token")
// 	}
// 	p.Token = token
// 	p.saveToken()
// 	return nil
// }
