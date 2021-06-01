package hidrive

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"text/template"

	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/auth"
)

func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

func (d *Drive) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") //r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	claims, _, err := auth.GetClaims(r)
	token, err := d.manager.GetAuthToken(claims.Username)
	fmt.Println("claims:", claims)
	if token == nil || err != nil {
		http.Error(w, errors.NewWithCode(401, "Couldn‘t get valid auth token for authuser %q", claims.Username).Error(), 401)
		return
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")

	head, tail := shiftPath(r.URL.Path)

	var meta *Meta
	var body io.ReadCloser

	switch head {

	case "dir":
		switch r.Method {
		case http.MethodGet:
			// meta, err = d.client.GetDir(d.clean(tail, token.Alias), "", "", 0, 0, "", "", token.AccessToken)
			meta, err = d.Listdir(tail, claims.Username)
			// meta, err = d.GetMeta(tail, claims.Username)

		case http.MethodPost:
			meta, err = d.Mkdir(tail, claims.Username)

		case http.MethodDelete:
			err = d.Rmdir(tail, claims.Username)
		}

	case "meta":
		switch r.Method {
		case http.MethodGet:
			meta, err = d.client.GetMeta(d.clean(tail, claims.Username), "", "", token.AccessToken)
		case http.MethodPut:
			meta, err = d.Save(tail, r.Body, claims.Username)
		}

	case "files":
		body, err = d.client.GetFile("", tail[1:], token.AccessToken)
		if err == nil {
			defer body.Close()
			_, err = io.Copy(w, body)
		}

	case "serve":
		body, err = d.client.GetFile(d.clean(tail, token.Alias), "", token.AccessToken)
		if err == nil {
			defer body.Close()
			_, err = io.Copy(w, body)
		}

	case "thumbs":
		params := r.URL.Query()
		if len(tail) > 1 {
			params.Set("path", d.clean(tail, token.Alias))
		}
		body, err = d.client.Request("GET", "/file/thumbnail", params, nil, token.AccessToken) //url.Values{"pid": {tail[1:]}})
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
		fmt.Println("default:", r.URL.Path, claims.Username)
		meta, err = d.GetMeta(r.URL.Path, claims.Username)
	}

	if meta != nil {
		enc.Encode(meta)
	}
	if err != nil {
		code := errors.Code(err)
		if code == 65535 {
			code = 500
		}
		http.Error(w, err.Error(), code)
		return
	}

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
	claims, _, _ := auth.GetClaims(r)
	token, _ := d.manager.GetAuthToken(claims.Username)
	fmt.Println("Serve", claims, token)
	params := r.URL.Query()
	params.Set("path", d.clean(r.URL.Path, token.Alias))
	body, err := d.client.Request("GET", "/file", params, nil, token.AccessToken)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer body.Close()
	_, err = io.Copy(w, body)

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
		m.Refresh(key)

	default:

		if key, found := r.URL.Query()["refresh"]; found {
			m.Refresh(key[0])
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
