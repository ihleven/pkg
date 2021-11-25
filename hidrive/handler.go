package hidrive

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	gopath "path"
	"strings"
	"text/template"

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

	var (
		err  error
		meta *Meta
		body io.ReadCloser
	)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")

	head, path := shiftPath(r.URL.Path)

	switch head {

	case "meta":
		// fullpath := d.fullpath(r.URL.Path, token.Alias)
		switch r.Method {
		case http.MethodGet:
			meta, err = d.Meta(d.fullpath(path, token.Alias), token)
		case http.MethodPut:
			dir, file := gopath.Split(d.fullpath(path, token.Alias))
			dir = strings.TrimSuffix(dir, "/")
			meta, err = d.client.PutFile(body, dir, file, 0, 0, token.AccessToken)
			// meta, err = d.Save(tail, r.Body, claims.Username)
		}

	case "dir":
		fullpath := d.fullpath(r.URL.Path, token.Alias)

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

	case "files":

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

	case "info":
		info, err := d.manager.oauthClient.TokenInfo(token.AccessToken)
		if err == nil {
			enc.Encode(info)
			return
		}

	case "authorize":
		clientAuthURL := d.manager.GetClientAuthorizeURL(path[1:], r.URL.Query().Get("next"))
		http.Redirect(w, r, clientAuthURL, 302)
		fmt.Fprintf(w, "%s", clientAuthURL)
		fmt.Printf("%s", clientAuthURL)

		// https://my.hidrive.com/client/authorize
		// ?client_id=a7ff922a897cfde20473f1b8d01b42a9
		// &lang=de
		// &redirect_uri=http%3A%2F%2Flocalhost%3A8000%2Fhidrive%2Fauth%2Fauthcode
		// &response_type=code
		// &scope=user%2Crw
		// &state=#login

	case "auth":
		// d.manager.ServeHTTP(w, r)
		head, tail := shiftPath(path)

		switch head {

		case "authorize":
			clientAuthURL := d.manager.GetClientAuthorizeURL(strings.TrimPrefix(tail, "/"), r.URL.Query().Get("next"))
			enc.Encode(map[string]string{"url": clientAuthURL})
			// http.Redirect(w, r, clientAuthURL, 302)

		case "authcode":
			// HandleAuthorizeCallback verarbeitet die Weiterleitung nach erfolgtem authorize
			// sollte unter  https://ihle.cloud/hidrive-token-callback erreichbar sein
			key := r.URL.Query().Get("state")
			fmt.Println("authcode:", key)
			if code, ok := r.URL.Query()["code"]; ok {
				fmt.Println("code", code, ok)
				_, err := d.manager.AddAuth(key, code[0])
				if err != nil {
					fmt.Printf("error => %+v\n", err)
					fmt.Fprintf(w, "error => %+v\n", err)
				}
				// p.writeTokenFile()

				next := r.URL.Query().Get("next")
				if next == "" {
					next = "/hidrive/auth"
				}
				fmt.Println("code", code, ok, next)
				fmt.Fprintf(w, `<head><meta http-equiv="Refresh" content="0; URL=%s"></head>`, next)

			}
		case "refresh":
			key := strings.TrimPrefix(tail, "/")
			d.manager.Refresh(key, false)

		default:

			if key, found := r.URL.Query()["refresh"]; found {
				d.manager.Refresh(key[0], false)
			}

			// t, err := template.ParseFS(templates, "hdhandler/templates/*.html")
			// if err != nil {
			// 	fmt.Fprintf(w, "Cannot parse templates: %v", err)
			// 	return
			// }
			enc.Encode(map[string]interface{}{
				"ClientID":     d.manager.clientID,
				"ClientSecret": d.manager.clientSecret,
				"tokens":       d.manager.authmap,
			})

		}
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

//go:embed templates/*
var templates embed.FS

// ServeHTTP has 3 modes:
// 1) GET without params shows a button calling ServeHTTP with the authorize param
//    and a form POSTing ServeHTTP with a code
// 2) GET with authorize param redirects to the hidrive /client/authorize endpoint where a user can authorize the app.
//    This endpoint will call registered token-callback which is not reachable locally ( => copy code and use form from 1)
// 3) POSTing username and code triggering oauth2/token endpoint generating an access token
func (m *AuthManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")

	head, tail := shiftPath(r.URL.Path)

	switch head {

	case "authorize":
		// clientAuthURL := m.GetClientAuthorizeURL(r.URL.Query().Get("state"), r.URL.Query().Get("next"))
		key := strings.TrimPrefix(tail, "/")
		clientAuthURL := m.GetClientAuthorizeURL(key, r.URL.Query().Get("next"))
		http.Redirect(w, r, clientAuthURL, 302)

	case "authcode":
		// HandleAuthorizeCallback verarbeitet die Weiterleitung nach erfolgtem authorize
		// sollte unter  https://ihle.cloud/hidrive-token-callback erreichbar sein
		key := r.URL.Query().Get("state")
		fmt.Println("authcode:", key)
		if code, ok := r.URL.Query()["code"]; ok {
			fmt.Println("code", code, ok)
			_, err := m.AddAuth(key, code[0])
			if err != nil {
				fmt.Printf("error => %+v\n", err)
				fmt.Fprintf(w, "error => %+v\n", err)
			}
			// p.writeTokenFile()

			next := r.URL.Query().Get("next")
			if next == "" {
				next = "/hidrive/auth"
			}
			fmt.Println("code", code, ok, next)
			fmt.Fprintf(w, `<head><meta http-equiv="Refresh" content="0; URL=%s"></head>`, next)

		}
	case "refresh":
		key := strings.TrimPrefix(tail, "/")
		m.Refresh(key, false)

	default:

		if key, found := r.URL.Query()["refresh"]; found {
			m.Refresh(key[0], false)
		}

		t, err := template.ParseFS(templates, "hdhandler/templates/*.html")
		if err != nil {
			fmt.Fprintf(w, "Cannot parse templates: %v", err)
			return
		}

		w.WriteHeader(200)
		t.ExecuteTemplate(w, "tokenmgmt.html", map[string]interface{}{"ClientID": m.clientID, "ClientSecret": m.clientSecret, "tokens": m.authmap})
		return
	}
}

func (m *AuthManager) AuthTokenCallback(w http.ResponseWriter, r *http.Request) {

	key := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	_, err := m.AddAuth(key, code)
	if err != nil {
		fmt.Printf("error => %+v\n", err)
		fmt.Fprintf(w, "error => %+v\n", err)
	}
	// p.writeTokenFile()

	next := r.URL.Query().Get("next")
	if next == "" {
		next = "/hidrive/auth"
	}

	fmt.Fprintf(w, `<head><meta http-equiv="Refresh" content="0; URL=%s"></head>`, next)

}
