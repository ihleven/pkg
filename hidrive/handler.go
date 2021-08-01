package hidrive

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	gopath "path"
	"strings"
	"sync"
	"text/template"

	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/auth"
)

func shiftPath(p string) (head, tail string) {
	p = gopath.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

func (d *Drive) credentials(r *http.Request) (string, string) {
	fmt.Println("credentials")
	claims, _, err := auth.GetClaims(r)
	if err != nil {
		fmt.Println("MauthManager.credentials -1- ERROR:", err)
		return "", ""
	}
	fmt.Println("claims:", claims)
	token, err := d.manager.GetAuthToken(claims.Username)
	if err != nil {
		fmt.Println("MauthManager.credentials - 2 - ERROR:", err)
		return "", ""
	}
	return d.clean(r.URL.Path, token.Alias), token.AccessToken
}

// ServeHTTP implements http.Handler interface for a drive.
//
func (d *Drive) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") //r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	path, accessToken := d.credentials(r)
	if accessToken == "" {
		http.Error(w, errors.NewWithCode(401, "Couldn‘t get valid auth token").Error(), http.StatusUnauthorized)
		return
	}

	fmt.Println("drive.Handler", r.URL.Path, path, accessToken)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")

	head, tail := shiftPath(r.URL.Path)

	var err error
	var meta *Meta
	var body io.ReadCloser

	switch head {

	case "dir":
		switch r.Method {
		case http.MethodGet:
			meta, err = d.client.GetDir(path, "", "", 0, 0, "", "", accessToken)
			// meta, err = d.Listdir(tail, claims.Username)

		case http.MethodPost:
			meta, err = d.client.PostDir(path, "", "", 0, 0, accessToken)
			// meta, err = d.Mkdir(path, accessToken)

		case http.MethodDelete:
			params := url.Values{"path": {path}, "recursive": {"true"}}
			_, err = d.client.Request("DELETE", "/dir", params, nil, accessToken)
			// err = d.Rmdir(tail, claims.Username)
		}

	case "meta":
		switch r.Method {
		case http.MethodGet:
			meta, err = d.client.GetMeta(path, "", "", accessToken)
		case http.MethodPut:
			dir, file := gopath.Split(path)
			dir = strings.TrimSuffix(dir, "/")
			meta, err = d.client.PutFile(body, dir, file, 0, 0, accessToken)
			// meta, err = d.Save(tail, r.Body, claims.Username)
		}

	case "files":
		pid := tail[1:]
		body, err = d.client.GetFile("", pid, accessToken)
		if err == nil {
			defer body.Close()
			_, err = io.Copy(w, body)
		}

	case "serve":
		body, err = d.client.GetFile(path, "", accessToken)
		if err == nil {
			defer body.Close()
			_, err = io.Copy(w, body)
		}

	case "thumbs":
		params := r.URL.Query()
		if len(tail) > 1 {
			params.Set("path", path)
		}
		body, err = d.client.Request("GET", "/file/thumbnail", params, nil, accessToken) //url.Values{"pid": {tail[1:]}})
		if err == nil {
			defer body.Close()
			_, err = io.Copy(w, body)
		}

	case "authorize":
		clientAuthURL := d.manager.GetClientAuthorizeURL(tail[1:], r.URL.Query().Get("next"))
		http.Redirect(w, r, clientAuthURL, 302)
		fmt.Fprintf(w, "%s", clientAuthURL)
		fmt.Printf("%s", clientAuthURL)

	default:
		var claims auth.Claims
		claims, _, err = auth.GetClaims(r)
		fmt.Println("drive handler default:", claims, r.URL.Path, claims.Username)
		meta, err = d.GetMeta(r.URL.Path, claims.Username)
	}

	switch {

	case err != nil:
		code := errors.Code(err)
		if code == 65535 {
			code = 500
		}
		http.Error(w, err.Error(), code)

	case meta != nil:
		enc.Encode(meta)
	}
}
func (d *Drive) MetaHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") //r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	path, accessToken := d.credentials(r)
	if accessToken == "" {
		http.Error(w, errors.NewWithCode(401, "Couldn‘t get valid auth token").Error(), http.StatusUnauthorized)
		return
	}

	fmt.Println("drive.MetaHandler", r.URL.Path, path, accessToken)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")

	// head, tail := shiftPath(r.URL.Path)

	var err error
	var meta *Meta

	var wg sync.WaitGroup
	var dir *Meta
	var direrr error

	wg.Add(1)

	go func() {
		defer wg.Done()
		dir, direrr = d.client.GetDir(path, "", "", 0, 0, "", "", accessToken)
	}()

	meta, err = d.client.GetMeta(path, "", "", accessToken)
	if err != nil {
		http.Error(w, errors.Wrap(err, "Couldn‘t get meta").Error(), 500)
		return
	}
	if meta.Path == "/users/matt.ihle" {
		meta.Path = "/"
		meta.NameURLEncoded = "home"
	}
	if strings.HasPrefix(meta.Path, "/users/matt.ihle") {
		meta.Path = strings.TrimPrefix(meta.Path, "/users/matt.ihle")
	}

	if meta.Filetype == "dir" {
		wg.Wait()
		if direrr != nil {
			http.Error(w, errors.Wrap(err, "Couldn‘t get dir").Error(), 500)
			return
		}
		meta.Members = dir.Members
		for i, m := range meta.Members {
			// fmt.Println("meta handler -> dir", m.NameURLEncoded, m.Name())

			if meta.Members[i].Path == "/users/matt.ihle" {
				meta.Members[i].Path = "/"
			}
			if strings.HasPrefix(meta.Members[i].Path, "/users/matt.ihle") {
				meta.Members[i].Path = strings.TrimPrefix(meta.Members[i].Path, "/users/matt.ihle")
			}
			meta.Members[i].NameURLEncoded = m.Name()
			// unescapedName, _ := url.QueryUnescape(m.NameURLEncoded)
			// m.NameURLEncoded = unescapedName
		}
	}

	switch {

	case err != nil:
		code := errors.Code(err)
		if code == 65535 {
			code = 500
		}
		http.Error(w, err.Error(), code)

	case meta != nil:
		enc.Encode(meta)
	}
}

func (d *Drive) DirHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") //r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	path, accessToken := d.credentials(r)
	if accessToken == "" {
		http.Error(w, errors.NewWithCode(401, "Couldn‘t get valid auth token").Error(), http.StatusUnauthorized)
		return
	}

	fmt.Println("drive.DirHandler", r.URL.Path, path, accessToken)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")

	// head, tail := shiftPath(r.URL.Path)

	var err error
	var meta *Meta

	switch r.Method {
	case http.MethodGet:
		meta, err = d.client.GetDir(path, "", "", 0, 0, "", "", accessToken)

	case http.MethodPost:
		meta, err = d.client.PostDir(path, "", "", 0, 0, accessToken)

	case http.MethodDelete:
		params := url.Values{"path": {path}, "recursive": {"true"}}
		_, err = d.client.Request("DELETE", "/dir", params, nil, accessToken)
	}

	switch {

	case err != nil:
		code := errors.Code(err)
		if code == 65535 {
			code = 500
		}
		http.Error(w, err.Error(), code)

	case meta != nil:
		for _, m := range meta.Members {
			// m.Path = ""
			m.NameURLEncoded = m.Name()
			unescapedName, _ := url.QueryUnescape(m.NameURLEncoded)
			m.NameURLEncoded = unescapedName
		}
		enc.Encode(meta)
	}
}

func (d *Drive) ThumbHandler(w http.ResponseWriter, r *http.Request) {
	path, accessToken := d.credentials(r)
	params := r.URL.Query()
	params.Set("path", path)
	//url.Values{"pid": {tail[1:]}})

	body, err := d.client.Request("GET", "/file/thumbnail", params, nil, accessToken)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer body.Close()
	_, err = io.Copy(w, body)
}

func (d *Drive) Serve(w http.ResponseWriter, r *http.Request) {

	path, accessToken := d.credentials(r)

	params := r.URL.Query()
	params.Set("path", path)

	body, err := d.client.Request("GET", "/file", params, nil, accessToken)
	if err != nil {
		http.Error(w, err.Error(), errors.Code(err))
		return
	}
	defer body.Close()

	_, err = io.Copy(w, body)
	if err != nil {
		http.Error(w, err.Error(), errors.Code(err))
	}
}

//go:embed hdhandler/templates/*
var templates embed.FS

func (m *AuthManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// claims, _, err := auth.GetClaims(r)
	// token, err := m.GetAuthToken(claims.Username)

	// if token == nil || err != nil {
	// 	http.Error(w, errors.NewWithCode(401, "Couldn‘t get valid auth token for authuser %q", claims.Username).Error(), 401)
	// 	return
	// }

	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")

	head, tail := shiftPath(r.URL.Path)

	switch head {

	case "authorize":
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
