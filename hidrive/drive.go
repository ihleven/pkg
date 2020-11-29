package hidrive

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/ihleven/errors"
)

// NewDrive creates a new hidrive
func NewDrive(oap *OAuth2Prov, pathfunc func(*http.Request) (string, string, string)) *Drive {

	var Drive = Drive{
		client:   NewClient(oap, ""),
		pathfunc: pathfunc,
	}
	return &Drive
}

type Drive struct {
	client   *HiDriveClient
	pathfunc func(*http.Request) (string, string, string)
}

func PrefixPath(prefix string) func(*http.Request) (string, string, string) {
	return func(r *http.Request) (string, string, string) {
		head, tail := shiftPath(r.URL.Path)

		return head, tail, prefix
	}
}

// func (d *PrefixDrive) rootpath(r *http.Request) (string, string, error) {
// 	claims, status, err := auth.GetClaims(r)
// 	if status != 0 || err != nil {
// 		// w.WriteHeader(401)
// 		return errors.NewWithCode(401, "Claims error")
// 	}
// 	head, tail := shiftPath(r.URL.Path)
// homes: map[string]string{"matt": "/users/matt.ihle"}
// 	return head, path.Join(d.prefix, tail)
// }

func (d *Drive) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	head, tail, prefix := d.pathfunc(r)

	switch {
	case r.URL.Path == "/login":
		http.Redirect(w, r, "https://my.hidrive.com/client/authorize?client_id=b4436f1157043c2bf8db540c9375d4ed&response_type=code&scope=admin,rw", 302)

	case head == "dir":

		dir, err := d.client.GetDir(path.Join(prefix, tail), nil)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		bytes, _ := json.MarshalIndent(dir, "", "    ")
		w.Write(bytes)

	case head == "meta":
		meta, err := d.client.GetMeta(path.Join(prefix, tail))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		bytes, err := json.MarshalIndent(meta, "", "    ")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(bytes)

	case head == "files":
		params := r.URL.Query()
		if len(tail) > 1 {
			params["pid"] = []string{tail[1:]}
		}
		body, err := d.client.GetReadCloser("/file", params)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer body.Close()
		if _, err := io.Copy(w, body); err != nil {
			http.Error(w, errors.Wrap(err, "failed to Copy hidrive file to responseWriter").Error(), 500)
			return
		}
	case head == "file":
		params := r.URL.Query()
		if len(tail) > 1 {
			params["path"] = []string{path.Join(prefix, tail)}
		}
		body, err := d.client.GetReadCloser("/file", params)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer body.Close()
		if _, err := io.Copy(w, body); err != nil {
			http.Error(w, errors.Wrap(err, "failed to Copy hidrive file to responseWriter").Error(), 500)
			return
		}
	// case head == "uploadfile":
	// 	body, err := d.client.GetReadCloser("/file", r.URL.Query())
	// 	if err != nil {
	// 		http.Error(w, err.Error(), 500)
	// 		return
	// 	}
	// 	defer body.Close()
	// 	if _, err := io.Copy(w, body); err != nil {
	// 		http.Error(w, errors.Wrap(err, "failed to Copy hidrive file to responseWriter").Error(), 500)
	// 		return
	// 	}
	case head == "thumbs":
		params := r.URL.Query()
		if len(tail) > 1 {
			params["path"] = []string{path.Join(prefix, tail)}
		}
		body, err := d.client.GetReadCloser("/file/thumbnail", params) //url.Values{"pid": {tail[1:]}})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer body.Close()
		if _, err := io.Copy(w, body); err != nil {
			http.Error(w, errors.Wrap(err, "failed to Copy hidrive file to responseWriter").Error(), 500)
			return
		}

	// case strings.HasSuffix(r.URL.Path, "/") || path.Ext(r.URL.Path) == "":
	// 	fmt.Println("dir2:", r.URL.Path)
	// 	dir, err := d.client.GetDir(r.URL.Path)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		http.Error(w, err.Error(), 500)
	// 		return
	// 	}
	// 	bytes, _ := json.MarshalIndent(dir, "", "    ")
	// 	w.Write(bytes)

	default:

		body, err := d.client.GetReadCloser("/file", url.Values{"path": {r.URL.Path}})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if _, err := io.Copy(w, body); err != nil {
			http.Error(w, errors.Wrap(err, "failed to Copy hidrive file to responseWriter").Error(), 500)
			return
		}

	}

}

func (d *Drive) ServeMeta(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-control", "no-cache")
	w.Header().Set("Cache-control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	meta, err := d.client.GetMeta(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// bytes, err := json.MarshalIndent(meta, "", "    ")
	// if err != nil {
	// 	return err
	// }

	// tm := time.Unix(meta.MTime, 0)
	var response interface{}

	switch meta.Type {
	case "dir":
		response, err = d.client.GetDir(r.URL.Path, nil)
		if err != nil {
			if hderr, ok := err.(*HiDriveError); ok {
				// i, _ := strconv.Atoi(hderr.Code)
				http.Error(w, hderr.EMessage, hderr.ECode)
			}
			return
		}
	case "file":
		response = meta
	}
	bytes, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	w.Write(bytes)
	return
}

func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}
