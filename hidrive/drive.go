package hidrive

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/ihleven/errors"
)

// NewDrive creates a new hidrive
func NewDrive(oap *OAuth2Prov, pathfunc func(*http.Request) (string, string, string)) *Drive {

	var Drive = Drive{
		Auth:     oap,
		client:   NewClient(oap, ""),
		prefix:   "/wolfgang-ihle",
		pathfunc: pathfunc,
	}
	return &Drive
}

type Drive struct {
	client   *HiDriveClient
	Auth     *OAuth2Prov
	prefix   string
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

	// w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Cache-control", "no-cache")
	// w.Header().Set("Cache-control", "no-store")
	// w.Header().Set("Pragma", "no-cache")
	// w.Header().Set("Expires", "0")

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	head, tail, prefix := d.pathfunc(r)

	switch {

	case head == "dir":

		dir, err := d.GetDir(path.Join(prefix, tail), "")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		bytes, _ := json.MarshalIndent(dir, "", "    ")
		w.Write(bytes)

	case head == "meta":
		meta, err := d.GetMeta(tail, "")
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
		body, err := d.client.Request("GET", "/file", params, nil, "")
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
		body, err := d.client.Request("GET", "/file", params, nil, "")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer body.Close()
		if _, err := io.Copy(w, body); err != nil {
			http.Error(w, errors.Wrap(err, "failed to Copy hidrive file to responseWriter").Error(), 500)
			return
		}

	case head == "thumbs":
		params := r.URL.Query()
		if len(tail) > 1 {
			params["path"] = []string{path.Join(prefix, tail)}
		}
		body, err := d.client.Request("GET", "/file/thumbnail", params, nil, "") //url.Values{"pid": {tail[1:]}})
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

		body, err := d.client.Request("GET", "/file", url.Values{"path": {r.URL.Path}}, nil, "")
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

func (d *Drive) FullPath(path string, authuser string) string {
	return d.prefix + path // replace home by homedirpath
}

func (d *Drive) GetDir(path string, authuser string) (*DirResponse, error) {

	token, err := d.Auth.GetAccessToken(authuser)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t get valid auth token for authuser %q", authuser)
	}

	hidrivepath := d.prefix + path // replace home by homedirpath

	params := make(map[string][]string)

	memberfields := "members,members.id,members.name,members.nmembers,members.size,members.type,members.mime_type,members.mtime,members.image.height,members.image.width,members.image.exif"
	params["path"] = []string{hidrivepath}
	params["members"] = []string{"all"}
	params["fields"] = []string{metafields + "," + memberfields}

	body, err := d.client.Request("GET", "/dir", params, nil, token)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var response DirResponse
	err = json.NewDecoder(body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (d *Drive) GetMeta(pfad string, authuser string) (*Meta, error) {
	token, err := d.Auth.GetAccessToken(authuser)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t get valid auth token for authuser %q", authuser)
	}

	hidrivepath := path.Join(d.prefix, pfad) // replace home by homedirpath

	params := url.Values{
		"path":   {hidrivepath},
		"fields": {metafields + "," + imagefields},
	}
	body, err := d.client.Request("GET", "/meta", params, nil, token)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var meta Meta
	err = json.NewDecoder(body).Decode(&meta)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't decode response body")
	}

	return &meta, nil
	// fmt.Println("GetMEta:", meta)

	// switch meta.Type {
	// case "file":
	// 	return &meta, nil

	// case "dir":
	// 	dir, err := d.GetDir(pfad, authuser)
	// 	if err != nil {
	// 		if hderr, ok := err.(*HiDriveError); ok {
	// 			return nil, errors.Wrap(hderr, "")
	// 		}
	// 		return nil, errors.Wrap(err, "")
	// 	}
	// 	return dir, nil
	// }
	// return nil, errors.New("neither file nor dir")
}

func (d *Drive) File(pfad string, username string) (io.ReadCloser, error) {
	token, err := d.Auth.GetAccessToken(username)
	if err != nil {
		return nil, err
	}
	query := make(url.Values)
	if len(pfad) > 1 {
		query["path"] = []string{path.Join(d.prefix, pfad)}
	}
	body, err := d.client.Request("GET", "/file", query, nil, token)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return body, nil
}

func (d *Drive) Thumbnail(pfad string, query url.Values, username string) (io.ReadCloser, error) {
	token, err := d.Auth.GetAccessToken(username)
	if err != nil {
		return nil, err
	}
	if len(pfad) > 1 {
		query["path"] = []string{path.Join(d.prefix, pfad)}
	}
	body, err := d.client.Request("GET", "/file/thumbnail", query, nil, token)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return body, nil
}

func (d *Drive) Mkdir(dirname, basename, authuser string) (interface{}, error) {

	token, err := d.Auth.GetAccessToken(authuser)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t get valid auth token for authuser %q", authuser)
	}

	// type response struct {
	// 	Ctime    string    //    - timestamp - ctime of the directory
	// 	HasDirs  bool      //    - bool      - does the directory contain subdirs?
	// 	ID       string    //    - string    - path id of the directory
	// 	Mtime    time.Time //    - timestamp - mtime of the directory
	// 	Name     string    //    - string    - URL-encoded name of the directory
	// 	Path     string    //    - string    - URL-encoded path to the directory
	// 	Readable bool      //    - bool      - read-permission for the directory
	// 	Type     string    //    - string    - "dir"
	// 	Writable bool      //    - bool      - write-permission for the directory
	// }

	params := url.Values{
		"path": {path.Join(d.prefix, dirname, basename)},
		// "on_exist": {"autoname"},
	}

	body, err := d.client.Request("POST", "/dir", params, nil, token)
	if err != nil {
		return nil, errors.Wrap(err, "Error in post request")
	}
	defer body.Close()

	var dir interface{}
	err = json.NewDecoder(body).Decode(&dir)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding post result")
	}
	return &dir, nil
}
func (d *Drive) Rmdir(dirname, authuser string) error {

	token, err := d.Auth.GetAccessToken(authuser)
	if err != nil {
		return errors.Wrap(err, "Couldn‘t get valid auth token for authuser %q", authuser)
	}

	params := url.Values{
		"path":      {path.Join(d.prefix, dirname)},
		"recursive": {"true"},
	}
	_, err = d.client.Request("DELETE", "/dir", params, nil, token)
	if err != nil {
		return errors.Wrap(err, "Error in post request")
	}

	return nil
}
func (d *Drive) DeleteFile(filename string, authuser string) error {
	token, err := d.Auth.GetAccessToken(authuser)
	if err != nil {
		return errors.Wrap(err, "Couldn‘t get valid auth token for authuser %q", authuser)
	}

	params := url.Values{
		"path": {path.Join(d.prefix, filename)},
	}
	_, err = d.client.Request("DELETE", "/file", params, nil, token)
	if err != nil {
		return errors.Wrap(err, "Error in delete request")
	}

	return nil
}

func (d *Drive) CreateFile(folder string, body io.Reader, name string, modtime string, authuser string) (*Meta, error) {

	token, err := d.Auth.GetAccessToken(authuser)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t get valid auth token for authuser %q", authuser)
	}
	respBody, err := d.client.Request("POST", "/file", url.Values{
		"dir":      {path.Join(d.prefix, folder)},
		"name":     {name},
		"on_exist": {"autoname"},
		"mtime":    {modtime},
	}, body, token)
	if err != nil {
		return nil, errors.Wrap(err, "Error in post request")
	}
	defer respBody.Close()

	bytes, err := ioutil.ReadAll(respBody)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading request body")
	}
	fmt.Println("createfile:", err, path.Join(d.prefix, folder), name, modtime, authuser, token, string(bytes))

	type Response struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Path     string `json:"path"`
		Type     string `json:"type"`
		CTime    int64  `json:"ctime"`
		MTime    int64  `json:"mtime"`
		HasDirs  bool   `json:"has_dirs"`
		Readable bool   `json:"readable"`
		Writable bool   `json:"writable"`
		MIMEType string `json:"mime_type"`
		Size     uint64 `json:"size"`
		Image    *struct {
			Height int `json:"height"`
			Width  int `json:"width"`
			Exif   struct {
				DateTimeOriginal string
				ExifImageHeight  string
				ExifImageWidth   string
				Orientation      string
			} `json:"exif"`
		} `json:"image"`
	}
	var r Response
	// err = json.NewDecoder(respBody).Decode(&meta)
	err = json.Unmarshal(bytes, &r)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding post result")
	}

	// meta.Meta.Image = &Image{Height: meta.Image.Height, Width: meta.Image.Width, Exif: Exif{DateTimeOriginal: meta.Image.Exif["DateTimeOriginal"].(string)}}

	return &Meta{r.ID, r.Name, r.Path, r.Type, r.CTime, r.MTime, r.HasDirs, r.Readable, r.Writable, r.MIMEType, r.Size, 0, "",
		&Image{Height: r.Image.Height, Width: r.Image.Width},
	}, nil
}
