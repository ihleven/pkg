package hidrive

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	gopath "path"
	"strings"

	"github.com/ihleven/errors"
)

func shiftPath(p string) (head, tail string) {
	p = gopath.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

func (d *Drive) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	token := d.Token(r.Context().Value("username").(string))
	if token == nil {
		http.Error(w, "Couldn‘t get valid auth token", http.StatusUnauthorized)
		return
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")

	var (
		meta *Meta
		body io.ReadCloser
		err  error
	)

	head, path := shiftPath(r.URL.Path)

	switch head {

	case "meta":
		// fullpath := d.fullpath(r.URL.Path, token.Alias)
		switch r.Method {
		case http.MethodGet:
			meta, err = d.Meta(path, token)

		case http.MethodPut:
			dir, file := gopath.Split(d.fullpath(path, token.Alias))
			dir = strings.TrimSuffix(dir, "/")
			meta, err = d.client.PutFile(body, dir, file, 0, 0, token.AccessToken)
			// meta, err = d.Save(tail, r.Body, claims.Username)
		}

	case "dir":
		fullpath := d.fullpath(path, token.Alias)

		switch r.Method {
		case http.MethodGet:
			meta, err = d.client.GetDir(fullpath, "", "", 0, 0, "", "", token.AccessToken)
			// meta, err = d.Listdir(tail, claims.Username)

		case http.MethodPost:
			meta, err = d.client.PostDir(fullpath, "", "", 0, 0, token.AccessToken)
			// meta, err = d.Mkdir(path, accessToken)

		case http.MethodDelete:
			params := url.Values{"path": {path}, "recursive": {"true"}}
			_, err = d.client.Request("DELETE", "/dir", params, nil, token.AccessToken)
			// err = d.Rmdir(tail, claims.Username)
		}

	case "file":
		switch r.Method {
		case http.MethodGet:
			body, err = d.client.Request("GET", "/file", url.Values{"path": {d.fullpath(path, token.Alias)}}, nil, token.AccessToken)
			if err == nil {
				defer body.Close()
				_, err = io.Copy(w, body)
			}

		case http.MethodPost:
			// create
			dir, name := gopath.Split(d.fullpath(path, token.Alias))
			params := url.Values{
				"dir_id":   r.URL.Query()["dir_id"],
				"dir":      {strings.TrimRight(dir, "/")},
				"name":     {name},
				"on_exist": {"autoname"}, // Find another name if the destination already exists.
				//        mtime, Type: int The modification time (mtime) of the file system target to be set after the operation.
				// parent_mtime, Type: int The modification time (mtime) of the file system target's parent folder to be set after the operation.
			}
			fmt.Println(params)
			body, err = d.client.Request("POST", "/file", params, r.Body, token.AccessToken)
			if err == nil {
				defer body.Close()
				meta, err = d.processMetaResponse(body)
			} else {
				fmt.Println("post file error:", err)
			}

		case "PUT":
			dir, file := gopath.Split(d.fullpath(path, token.Alias))
			body, err = d.client.Request("PUT", "/file", url.Values{"dir": {strings.TrimRight(dir, "/")}, "name": {file}}, r.Body, token.AccessToken)
			if err == nil {
				defer body.Close()
				meta, err = d.processMetaResponse(body)
			}

		// case "PATCH":

		case http.MethodDelete:
			err = d.Rm(path, token)
		}

	case "files":
		switch r.Method {
		case http.MethodPost:
			// create
			params := url.Values{
				"dir_id":   r.URL.Query()["dir_id"],
				"name":     {path[1:]},
				"on_exist": {"autoname"}, // Find another name if the destination already exists.
			}
			fmt.Println(params)
			body, err = d.client.Request("POST", "/file", params, r.Body, token.AccessToken)
			if err == nil {
				defer body.Close()
				meta, err = d.processMetaResponse(body)
			} else {
				fmt.Println("post file error:", err)
			}
		}

		body, err = d.client.Request("GET", "/file", url.Values{"pid": {path[1:]}}, nil, token.AccessToken)
		if err == nil {
			defer body.Close()
			_, err = io.Copy(w, body)
		}

	case "serve":
		body, err = d.client.Request("GET", "/file", url.Values{"path": {d.fullpath(path, token.Alias)}}, nil, token.AccessToken)
		if err == nil {
			defer body.Close()
			_, err = io.Copy(w, body)
		}

	case "thumbs":
		params := r.URL.Query() // width, height, mode & pid
		if len(path) > 1 {      // splitpath leifert mindestens "/"
			params.Set("path", d.fullpath(path, token.Alias))
		}

		body, err = d.client.Request("GET", "/file/thumbnail", params, nil, token.AccessToken)
		if err == nil {
			defer body.Close()
			_, err = io.Copy(w, body)
		}

	case "tokeninfo":
		info, err := d.manager.oauthClient.TokenInfo(token.AccessToken)
		if err == nil {
			enc.Encode(info)
			return
		}

	case "authorize":
		d.manager.AuthorizeURL(w, r)
		return

	default:

		// meta, err = d.Stream(r.URL.Path, token)
		meta, err = d.GetMeta(r.URL.Path, token)
	}

	switch {

	case err != nil:
		code := errors.Code(err)
		if code == 65535 {
			code = 500
		}
		http.Error(w, err.Error(), code)

	case meta != nil:
		// for _, m := range meta.Members {
		// 	// m.Path = ""
		// 	m.NameURLEncoded = m.Name()
		// 	unescapedName, _ := url.QueryUnescape(m.NameURLEncoded)
		// 	m.NameURLEncoded = unescapedName
		// }
		enc.Encode(meta)
	}
}

func (d *Drive) ThumbHandler(w http.ResponseWriter, r *http.Request) {

	token := d.Token(r.Context().Value("username").(string))
	if token == nil {
		http.Error(w, "Couldn‘t get valid auth token", http.StatusUnauthorized)
		return
	}

	params := r.URL.Query()
	params.Set("path", d.fullpath(r.URL.Path, token.Alias))

	body, err := d.client.Request("GET", "/file/thumbnail", params, nil, token.AccessToken)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer body.Close()

	_, err = io.Copy(w, body)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (d *Drive) Serve(w http.ResponseWriter, r *http.Request) {

	token := d.Token(r.Context().Value("username").(string))
	if token == nil {
		http.Error(w, "Couldn‘t get valid auth token", http.StatusUnauthorized)
		return
	}

	params := r.URL.Query()
	params.Set("path", d.fullpath(r.URL.Path, token.Alias))

	body, err := d.client.Request("GET", "/file", params, nil, token.AccessToken)
	if err != nil {
		http.Error(w, err.Error(), errors.Code(err))
		return
	}
	defer body.Close()

	_, err = io.Copy(w, body)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (d *Drive) FileContent(w http.ResponseWriter, r *http.Request) {

	token := d.Token(r.Context().Value("username").(string))
	if token == nil {
		http.Error(w, "Couldn‘t get valid auth token", http.StatusUnauthorized)
		return
	}

	dir, file := path.Split(d.fullpath(r.URL.Path, token.Alias))

	params := r.URL.Query()
	params.Set("dir", strings.TrimRight(dir, "/"))
	params.Set("name", file)

	body, err := d.client.Request("PUT", "/file", params, r.Body, token.AccessToken)
	if err != nil {
		http.Error(w, err.Error(), errors.Code(err))
		return
	}
	defer body.Close()

	var meta Meta
	err = json.NewDecoder(body).Decode(&meta)
	if err != nil {
		http.Error(w, err.Error(), errors.Code(err))
		return
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	err = enc.Encode(meta)
	if err != nil {
		http.Error(w, err.Error(), errors.Code(err))
		return
	}
}

// ServeHTTP has 3 modes:
// 1) GET without params shows a button calling ServeHTTP with the authorize param
//    and a form POSTing ServeHTTP with a code
// 2) GET with authorize/{username} param redirects to the hidrive /client/authorize endpoint where a user {username} can authorize the app.
//    This endpoint will call registered token-callback which is not reachable locally ( => copy code and use form from 1)
// 3) POSTing username and code triggering oauth2/token endpoint generating an access token
func (m *AuthManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")

	head, tail := shiftPath(r.URL.Path)
	switch head {

	case "authorize":
		r.URL.Path = tail
		m.AuthorizeURL(w, r)
		return

	case "authcode":
		r.URL.Path = tail
		m.HandleAuthorizeCallback(w, r)
		return

	case "refresh":
		r.URL.Path = tail
		m.RefreshHandler(w, r)
		return

	case "keys":
		switch r.Method {
		case "DELETE":
			m.DelAuth(tail[1:])
			enc.Encode(m.authmap)
		}

	default:

		if key, found := r.URL.Query()["refresh"]; found {
			m.Refresh(key[0], false)
		}

		enc.Encode(m.authmap)
		return
	}
}

func (m *AuthManager) AuthorizeURL(w http.ResponseWriter, r *http.Request) {
	// https://my.hidrive.com/client/authorize
	// ?client_id=a7ff922a897cfde20473f1b8d01b42a9
	// &lang=de
	// &redirect_uri=http%3A%2F%2Flocalhost%3A8000%2Fhidrive%2Fauth%2Fauthcode
	// &response_type=code
	// &scope=user%2Crw
	// &state=#login

	username, _ := shiftPath(r.URL.Path)

	params := url.Values{
		"client_id":     {m.clientID},
		"response_type": {"code"},
		"scope":         {"user,rw"},
		"lang":          {"de"},     // optional: language in which the authorization page is shown
		"state":         {username}, // optional:
		"redirect_uri":  {"http://localhost:8000/hidrive/auth/authcode?next=" + r.URL.Query().Get("next")},
	}

	w.Write([]byte("https://my.hidrive.com/client/authorize?" + params.Encode()))

	// enc := json.NewEncoder(w)
	// enc.SetIndent("", "    ")
	// enc.Encode(map[string]string{
	// 	"username": username,
	// 	"url":      m.GetClientAuthorizeURL(username),
	// })
}

func (m *AuthManager) HandleAuthorizeCallback(w http.ResponseWriter, r *http.Request) {

	key := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	next := r.URL.Query().Get("next")

	_, err := m.AddAuth(key, code)
	if err != nil {
		fmt.Printf("error => %+v\n", err)
		fmt.Fprintf(w, "error => %+v\n", err)
	}

	if next == "" {
		next = "/hidrive/auth"
	}
	// w.Write([]byte(next))
	fmt.Fprintf(w, `<head><meta http-equiv="Refresh" content="0; URL=%s"></head>`, next)
}

func (m *AuthManager) RefreshHandler(w http.ResponseWriter, r *http.Request) {

	authkey, _ := shiftPath(r.URL.Path)
	fmt.Println("authkey:", authkey)
	token, err := m.Refresh(authkey, false)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(token)
}
