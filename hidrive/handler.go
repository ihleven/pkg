package hidrive

import (
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

func HiDriveHandler(drive *Drive) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// w.Header().Set("Access-Control-Allow-Origin", "*")
		// w.Header().Set("Cache-control", "no-cache")
		// w.Header().Set("Cache-control", "no-store")
		// w.Header().Set("Pragma", "no-cache")
		// w.Header().Set("Expires", "0")

		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		var username string
		claims, _, err := auth.GetClaims(r)
		if err == nil {
			username = claims.Username
		}

		// head, tail, prefix := drive.pathfunc(r)
		head, tail := shiftPath(r.URL.Path)

		var response Responder

		switch {

		case head == "dir":
			response, err = drive.GetDir(tail, username)

		case head == "file":
			params := r.URL.Query()
			if len(tail) > 1 {
				params["path"] = []string{path.Join(drive.prefix, tail)}
			}
			body, err := drive.client.Request("GET", "/file", params, nil, username)
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

			body, err := drive.client.Request("GET", "/file", url.Values{"path": {r.URL.Path}}, nil, username)
			if err == nil {
				_, err = io.Copy(w, body)
			}

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
