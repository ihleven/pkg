package hidrive

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/ihleven/errors"
)

// AppConfig contains static app configuration
type AppConfig struct {
	ClientID     string `json:"-"`
	ClientSecret string `json:"-"`
}

// ReadFromFile parses app config from given file
func (c *AppConfig) ReadFromFile(filename string) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "Could not read file %q", filename)
	}
	err = json.Unmarshal(bytes, c)
	if err != nil {
		return errors.Wrap(err, "Could not unmarshal file %q", filename)
	}
	return nil
}

// NewOauthProvider constructs an auth provider
func NewOauthProvider(config AppConfig) (*OAuth2Prov, error) {

	// var config AppConfig
	// err := config.ReadFromFile(configFilePath)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "Couldn‘t read Config from file")
	// }

	oap := OAuth2Prov{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		tokenFile: "hidrive.tokens",
		// tokens:    make(map[string]Token),
		// deadlines: make(map[string]time.Time),
		cache:  make(map[string]*entry),
		tokens: make(map[string]*tokenmgmt),
	}

	err := oap.readTokenFile()
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t read Config from file")
	}

	return &oap, nil
}

func (p *OAuth2Prov) readTokenFile() error {
	bytes, err := ioutil.ReadFile(p.tokenFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Wrap(err, "failed to read token file")
	}
	tokens := make(map[string]*Token)
	if err := json.Unmarshal(bytes, &tokens); err != nil {
		return errors.Wrap(err, "error parsing token")
	}
	for key, token := range tokens {
		p.cache[key] = &entry{
			res: struct {
				token *Token
				err   error
			}{token: token},
		}
		p.tokens[key] = &tokenmgmt{
			Token: token,
		}
	}
	fmt.Printf("readTokenFile() => %v\n", p.cache)

	return nil
}

func (p *OAuth2Prov) writeTokenFile() error {
	tokens := make(map[string]*Token)
	for key, mgmt := range p.tokens {
		tokens[key] = mgmt.Token
	}
	bytes, err := json.MarshalIndent(tokens, "", "    ")
	if err != nil {
		return errors.Wrap(err, "error parsing token")
	}
	err = ioutil.WriteFile(p.tokenFile, bytes, 0664)
	if err != nil {
		return err
	}
	return nil
}

type OAuth2Prov struct {
	config    AppConfig    `json:"-"`
	client    *http.Client `json:"-"`
	tokenFile string       `json:"-"`
	// tokens    map[string]Token     `json:"-"`
	// deadlines map[string]time.Time `json:"-"`

	mu     sync.Mutex // guards cache
	cache  map[string]*entry
	tokens map[string]*tokenmgmt
}
type entry struct {
	res struct {
		token *Token
		err   error
	}
	deadline time.Time
	ready    chan struct{} // closed when res is ready
}
type tokenmgmt struct {
	*Token
	sync.RWMutex
	err      error
	deadline time.Time
	ready    chan struct{} // closed when res is ready
}

func (e *entry) minutesToDeadline() float64 {
	if e != nil {
		return e.deadline.Sub(time.Now()).Minutes()
	}
	return 0
}

func (p *OAuth2Prov) ForceRefresh(key string) error {
	p.mu.Lock()
	mgmt, ok := p.tokens[key]
	p.mu.Unlock()
	if !ok {
		return errors.New("no mgmt entry for key %s", key)
	}

	if mgmt == nil {
		return errors.New("empty mgmt entry for key %s", key)
	}
	newmgmt, err := p.RefreshMgmt(mgmt)
	if err != nil {
		return errors.New("Could not refresh mgmt %v", key)
	}
	p.tokens[key] = newmgmt

	return nil
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
		newmgmt, err := p.RefreshMgmt(mgmt)
		if err == nil {
			// fmt.Printf(" ** successful token refesh 1: %s -> %+v\n", key, newmgmt.Token)

			p.tokens[key] = newmgmt
			mgmt = newmgmt
			remaining = mgmt.deadline.Sub(time.Now()).Minutes()
			// fmt.Printf(" ** successful token refesh: %s -> %+v %v \n", key, mgmt.Token, mgmt.deadline)
			p.mu.Unlock()
			return mgmt.AccessToken, nil
		} else {
			p.mu.Unlock()
			// TODO: how to handle error?
			return "", errors.Wrap(err, "unable to refresh")
		}
	}

	if remaining > 0 {
		if remaining < 10 {
			// fmt.Printf(" ***************** triggering refesh: %s -> %f **************\n", key, remaining)
			// refresh token in background
			go func(mgmt *tokenmgmt) {
				newmgmt, err := p.RefreshMgmt(mgmt)
				if err == nil {
					p.mu.Lock()
					p.tokens[key] = newmgmt
					mgmt = newmgmt
					remaining = mgmt.deadline.Sub(time.Now()).Minutes()
					// fmt.Printf(" ***************** successful early token refesh: %s -> %f **************\n", key, remaining)

					p.mu.Unlock()
				} else {
					// TODO: how to handle error?
					log.Fatal("failed token refesh")
				}
			}(mgmt)
		}
		return mgmt.AccessToken, nil
	}

	return "", errors.New("Could not retrieve access token for key %s", key)
}

// https://notes.shichao.io/gopl/ch9/
func (memo *OAuth2Prov) GetAccessToken2(key string) string { //(value interface{}, err error) {
	fmt.Printf("GetAccessToken(%s)\n", key)

	memo.mu.Lock()
	// fmt.Println("minutesToDeadline:", minutesToDeadline, memo.deadlines)

	e := memo.cache[key]
	// minutesToDeadline := memo.deadlines[key].Sub(time.Now()).Minutes()
	if e == nil {
		fmt.Printf("GetAccessToken(%s) -> e is nil\n", key)
		memo.mu.Unlock()
		return ""
	}
	if e.minutesToDeadline() < 59 {
		fmt.Println("refresh:", e.res.token.RefreshToken, e.minutesToDeadline())
		refreshToken := e.res.token.RefreshToken

		e = &entry{ready: make(chan struct{})}
		memo.cache[key] = e
		memo.mu.Unlock()
		fmt.Println("refresh:")

		token, err := memo.RefreshToken2(refreshToken)
		// e.res.token, e.res.err = memo.RefreshToken2(e.res.token.RefreshToken)
		fmt.Println("refresh:", e.res.token, e.res.err)
		if err != nil {
			log.Fatal(" => refrseh token error:", err)
			return ""
		}
		e.res.token = token
		e.deadline = time.Now().Add(time.Second * time.Duration(e.res.token.ExpiresIn))

		close(e.ready) // broadcast ready condition
	} else {
		fmt.Println("no refresh:", e, e.minutesToDeadline())
		// This is a repeat request for this key.
		memo.mu.Unlock()

		<-e.ready // wait for ready condition
	}
	fmt.Printf("GetAccessToken(%s) => %s\n", key, e.res.token.AccessToken)
	return e.res.token.AccessToken // e.res.token, e.res.err
}

func (p *OAuth2Prov) TokenWithCode(key, code string) (*Token, error) {

	resp, err := p.client.PostForm("https://my.hidrive.com/oauth2/token", url.Values{
		"client_id":     {p.config.ClientID},
		"client_secret": {p.config.ClientSecret},
		"grant_type":    {"authorization_code"},
		"code":          {code},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var token Token
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}
	// p.tokens[key] = token
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tokens[key] = &tokenmgmt{
		Token:    &token,
		deadline: time.Now().Add(time.Second * time.Duration(token.ExpiresIn)),
	}
	fmt.Println("TokenWithCode:", token, key, code)
	return &token, nil
}

// func (p *OAuth2Prov) RefreshToken(key string) (*Token, error) {

// 	fmt.Println("refreshToken waiting for 10 seconds")

// 	token, ok := p.tokens[key]
// 	if !ok {
// 		return nil, errors.New("token key %q not found", key)
// 	}

// 	resp, err := p.client.PostForm("https://my.hidrive.com/oauth2/token", url.Values{
// 		"client_id":     {p.config.ClientID},
// 		"client_secret": {p.config.ClientSecret},
// 		"grant_type":    {"refresh_token"},
// 		"refresh_token": {token.RefreshToken},
// 	})
// 	if err != nil {
// 		return nil, errors.Wrap(err, "failed to call https://my.hidrive.com/oauth2/token")
// 	}
// 	defer resp.Body.Close()

// 	err = json.NewDecoder(resp.Body).Decode(&token)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "failed to decode response")
// 	}
// 	p.tokens[key] = token

// 	return &token, nil
// }

func (p *OAuth2Prov) RefreshToken2(token string) (*Token, error) {
	fmt.Println("RefreshToken2", token)

	resp, err := p.client.PostForm("https://my.hidrive.com/oauth2/token", url.Values{
		"client_id":     {p.config.ClientID},
		"client_secret": {p.config.ClientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {token},
	})
	fmt.Println("asdf", err)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call https://my.hidrive.com/oauth2/token")
	}
	defer resp.Body.Close()
	var newtoken Token
	err = json.NewDecoder(resp.Body).Decode(&newtoken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return &newtoken, nil
}

func (p *OAuth2Prov) RefreshMgmt(mgmt *tokenmgmt) (*tokenmgmt, error) {
	fmt.Printf("RefreshMgmt: Token -> %+v \n", mgmt.Token)

	resp, err := p.client.PostForm("https://my.hidrive.com/oauth2/token", url.Values{
		"client_id":     {p.config.ClientID},
		"client_secret": {p.config.ClientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {mgmt.RefreshToken},
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to call https://my.hidrive.com/oauth2/token")
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.NewWithCode(errors.ErrorCode(resp.StatusCode), string(body))
	}

	var newtoken Token
	decerr := json.NewDecoder(resp.Body).Decode(&newtoken)
	if decerr != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}
	fmt.Printf("RefreshMgmt -> %+v %v \n", newtoken, time.Now().Add(time.Second*time.Duration(newtoken.ExpiresIn)))
	return &tokenmgmt{
		Token:    &newtoken,
		err:      err,
		deadline: time.Now().Add(time.Second * time.Duration(newtoken.ExpiresIn)),
	}, nil
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

// func (p *OAuth2Prov) TokenInfo(key string) (*TokenInfo, error) {

// 	token, ok := p.tokens[key]
// 	if !ok {
// 		return nil, errors.New("unknown key %q", key)
// 	}

// 	res, err := p.client.PostForm("https://my.hidrive.com/oauth2/tokeninfo", url.Values{"access_token": {token.AccessToken}})
// 	if err != nil {
// 		return nil, errors.Wrap(err, "Error in tokeninfo request")
// 	}
// 	defer res.Body.Close()

// 	var info TokenInfo
// 	err = json.NewDecoder(res.Body).Decode(&info)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "Couldn‘t parse tokeninfo response")
// 	}
// 	return &info, nil
// }

func (p OAuth2Prov) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if state := r.URL.Query().Get("authorize"); state != "" {

		params := url.Values{
			"client_id":     {p.config.ClientID},
			"response_type": {"code"},
			"scope":         {"user,rw"}, //
			"lang":          {"de"},      // optional: language in which the authorization page is shown
			"state":         {state},     // optional:
		}
		http.Redirect(w, r, "https://my.hidrive.com/client/authorize?"+params.Encode(), 302)
	}
	if r.Method == "POST" {
		token, err := p.TokenWithCode(r.FormValue("username"), r.FormValue("code"))
		if err == nil {
			fmt.Fprintf(w, "token => %+v\n", token)
		}

		p.writeTokenFile()
	}
	tpl1 := `<html>
	<link href="https://unpkg.com/tailwindcss@^2/dist/tailwind.min.css" rel="stylesheet">
		<body>
			<a href="auth?authorize=ihle">authorize</a>
			<form method="post">
				<label>
					Username:
					<input name="username" />
				</label>
				<label>
					Code:
					<input name="code"/>
				</label>
				<input type="submit" value="submit" />
			</form>
		</body>
	</html>`
	fmt.Fprintf(w, tpl1)

}

// HandleAuthorizeCallback verarbeitet die Weiterleitung nach erfolgtem authorize
// sollte unter  https://ihle.cloud/hidrive-token-callback erreichbar sein
func (p *OAuth2Prov) HandleAuthorizeCallback(w http.ResponseWriter, r *http.Request) {
	// e.g. ?scope=rw,user&code=R8Pt3WTfbREUbx6SHnPn

	if scope, ok := r.URL.Query()["scope"]; ok {
		fmt.Fprintf(w, "scope => %q\n", scope)
	}
	code, ok := r.URL.Query()["code"]
	key := r.URL.Query().Get("state")
	if ok {
		fmt.Fprintf(w, "code => %q\n", code)
		// err := p.tokenWithCode(code[0])
		token, err := p.TokenWithCode(key, code[0])
		if err == nil {
			fmt.Fprintf(w, "token => %+v\n", token)
		}
		p.writeTokenFile()
	}
}

// Token zeigt ein Eingabefeld für einen Authorize-Code und fordert damit ein Token an
func (p *OAuth2Prov) TokenPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		token, err := p.TokenWithCode("", r.FormValue("code"))
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
