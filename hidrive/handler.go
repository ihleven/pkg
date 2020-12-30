package hidrive

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func HidriveHandler(drive *Drive) http.HandlerFunc {

	usermap := map[string]string{
		"matt":     "$2a$14$4zu/jv7JO377BBg0k6upZ.Ul0jqO9enCBhHlAUyoKJrcySb8JMzW2",
		"wolfgang": "$2a$14$KWkdJOJLa4FkKHyXZ9xFceutb8qqkQ0V2Ue1.Ce9Rn0OD69.tDHHC",
	}

	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, token")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		var username string
		claims, _, err := auth.GetClaims(r)
		if err == nil {
			username = claims.Username
		}

		// head, tail, prefix := drive.pathfunc(r)
		head, tail := shiftPath(r.URL.Path)

		var response Responder

		switch head {
		case "auth":
			// hidrive authentifizierung
			// das sollte sich zur auth admin page ausweiten: * übersicht hd accounts, ‘nen account authentifizieren/löschen, lokale auth als matt notwendig um auf die Seite zu kommen
			drive.Auth.ServeHTTP(w, r)

		case "refresh":
			err = drive.Auth.ForceRefresh(username)

		case "signin":
			// lokale authentifizierung mit usernamen wolfgang / matt
			auth.SigninHandler(usermap)(w, r)
		case "welcome":
			// lokale authentifizierung mit usernamen wolfgang / matt
			auth.Welcome(w, r)

		case "signout":
			auth.SignoutHandler(w, r)

		case "dir":
			response, err = drive.GetDir(tail, username)
		case "meta":
			response, err = drive.GetMeta(tail, username)
		case "thumbs":

			body, err := drive.Thumbnail(tail, r.URL.Query(), username)
			if err != nil {
				break
			}
			defer body.Close()
			if _, err := io.Copy(w, body); err != nil {
				err = errors.Wrap(err, "failed to Copy hidrive file to responseWriter")
			}
		case "file":
			token, e := drive.Auth.GetAccessToken(username)
			if err != nil {
				err = e
			}
			params := r.URL.Query()
			if len(tail) > 1 {
				params["path"] = []string{path.Join(drive.prefix, tail)}
			}
			body, err := drive.client.Request("GET", "/file", params, nil, token)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			defer body.Close()
			if _, err := io.Copy(w, body); err != nil {
				http.Error(w, errors.Wrap(err, "failed to Copy hidrive file to responseWriter").Error(), 500)
				return
			}
		default:
			var token string
			token, err = drive.Auth.GetAccessToken(username)

			query := url.Values{"path": []string{drive.HidrivePath(r.URL.Path, username)}}

			var body io.ReadCloser
			body, err = drive.client.Request("GET", "/file", query, nil, token)
			if err == nil {
				_, err = io.Copy(w, body)
			} else {
				fmt.Println(err)
			}

			// body, err := drive.client.Request("GET", "/file", url.Values{"path": {r.URL.Path}}, nil, username)
			// if err == nil {
			// 	_, err = io.Copy(w, body)
			// }

		}

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if response != nil {

			response.Respond(w, "")
		}

	}
}
