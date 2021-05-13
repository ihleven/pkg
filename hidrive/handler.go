package hidrive

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

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

	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	claims, _, err := auth.GetClaims(r)
	token, err := d.manager.GetAuthToken(claims.Username)

	if token == nil || err != nil {
		http.Error(w, errors.NewWithCode(401, "Couldn‘t get valid auth token for authuser %q", claims.Username).Error(), 401)
		return
	}

	// conf, ok := d.confmap[claims.Username]
	// if !ok {
	// 	conf = config{"anonymous", "/", []string{}}
	// }

	head, tail := shiftPath(r.URL.Path)

	switch head {

	case "dir":
		switch r.Method {
		case http.MethodGet:
			dir, err := d.client.GetDir(d.clean(tail, token.Alias), "", "", 0, 0, "", "", token.AccessToken)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			json.NewEncoder(w).Encode(dir)

		case http.MethodPost:
			_, err = d.Mkdir(tail, claims.Username)

		case http.MethodDelete:
			err = d.Rmdir(tail, claims.Username)
		}
	case "meta":
		meta, err := d.client.GetMeta(d.clean(tail, claims.Username), "", "", token.AccessToken)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(meta)

		// switch meta.Type {
		// case "dir":
		// 	dir, err := d.GetDir(pfad, authuser)
		// 	return dir, nil
		// }

	case "files":
		body, err := d.client.GetFile("", tail[1:], token.AccessToken)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer body.Close()
		if _, err := io.Copy(w, body); err != nil {
			http.Error(w, errors.Wrap(err, "failed to Copy hidrive file to responseWriter").Error(), 500)
			return
		}
	case "serve":

		body, err := d.client.GetFile(d.clean(tail, token.Alias), "", token.AccessToken)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer body.Close()
		if _, err := io.Copy(w, body); err != nil {
			http.Error(w, errors.Wrap(err, "failed to Copy hidrive file to responseWriter").Error(), 500)
			return
		}

	case "thumbs":
		params := r.URL.Query()
		if len(tail) > 1 {
			params["path"] = []string{d.clean(tail, token.Alias)}
		}
		body, err := d.client.Request("GET", "/file/thumbnail", params, nil, token.AccessToken) //url.Values{"pid": {tail[1:]}})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer body.Close()
		if _, err := io.Copy(w, body); err != nil {
			http.Error(w, errors.Wrap(err, "failed to Copy hidrive file to responseWriter").Error(), 500)
			return
		}
	case "authorize":
		clientAuthURL := d.manager.GetClientAuthorizeURL(tail[1:], r.URL.Query().Get("next"))
		http.Redirect(w, r, clientAuthURL, 302)
		fmt.Fprintf(w, "%s", clientAuthURL)
		fmt.Printf("%s", clientAuthURL)

	default:
		meta, err := d.GetMeta(r.URL.Path, claims.Username)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(&meta)
		// body, err := d.client.GetFile(d.clean(r.URL.Path, token.Alias), "", token.AccessToken)
		// // body, err := d.client.Request("GET", "/file", url.Values{"path": {"/" + head + tail}}, nil, "")
		// if err != nil {
		// 	fmt.Println(err)
		// 	if hderr, ok := err.(*HidriveError); ok {
		// 		fmt.Println(d.clean(path.Join(r.URL.Path, "index.html"), token.Alias))
		// 		if hderr.ErrorCode == 666 {

		// 			body, err = d.client.GetFile(d.clean(path.Join(r.URL.Path, "index.html"), token.Alias), "", token.AccessToken)
		// 		}
		// 	}
		// 	if err != nil {
		// 		http.Error(w, err.Error(), 500)
		// 		return
		// 	}
		// }
		// if _, err := io.Copy(w, body); err != nil {
		// 	http.Error(w, errors.Wrap(err, "failed to Copy hidrive file to responseWriter").Error(), 500)
		// 	return
		// }

	}

}
