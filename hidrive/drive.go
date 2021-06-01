package hidrive

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"path"
	"strings"
	"sync"

	"github.com/ihleven/errors"
)

// NewDrive creates a new hidrive
func NewDrive(clientID, clientSecret string, opts ...DriveOption) *Drive {

	var d = &Drive{
		client:  NewClient(),
		manager: NewAuthManager(clientID, clientSecret),
		// prefix:  "",
		// homes:    map[string]string{"matt": "/users/matt.ihle"},
		confmap: make(map[string]config),
	}

	// Loop through each option
	for _, opt := range opts {
		opt(d)
	}

	return d
}

// Drive ist ein Wrapper um Client, der Pfadumrechnungen und
type Drive struct {
	client  *HiDriveClient
	manager *AuthManager // Auth     *OAuth2Prov
	prefix  string
	useHome bool
	// homes    map[string]string
	confmap map[string]config
}

type config struct {
	Username string
	Prefix   string
	ACL      []string
}

type DriveOption func(*Drive)

func Prefix(prefix string) DriveOption {
	return func(d *Drive) {
		d.prefix = prefix
	}
}

func FromHome() DriveOption {
	return func(d *Drive) {
		d.useHome = true
	}
}
func (d *Drive) AM() *AuthManager {
	return d.manager
}
func (d *Drive) clean(inpath string, username string) string {

	outpath := path.Clean(inpath)

	if d.prefix != "" {
		outpath = path.Join(d.prefix, outpath)
	} else if d.useHome {
		outpath = path.Join("/users", username, outpath)
	} else if username != "" {
		if strings.HasPrefix(outpath, "/home") {
			outpath = strings.Replace(outpath, "/home", "/users/"+username, 1)
		}
		if strings.HasPrefix(outpath, "/~") {
			outpath = strings.Replace(outpath, "/~", "/users/"+username, 1)
		}
	}
	// else {
	//     tail = strings.Replace(tail, "/home", "/", 1)
	// }
	return outpath
}

// func (d *Drive) token(authkey string) *AuthToken {

// 	token, err := d.manager.GetAuthToken(authkey)
// 	if err != nil {
// 		fmt.Printf("%#v\n", errors.Wrap(err, "Couldn‘t get valid auth token for authuser %q", authkey))
// 	}
// 	return token
// }
func (d *Drive) GetMeta(path string, authkey string) (*Meta, error) {

	token, err := d.manager.GetAuthToken(authkey)
	if token == nil {
		return nil, errors.NewWithCode(401, "no valid token")
	}

	path = d.clean(path, token.Alias)

	var wg sync.WaitGroup
	var dir *Meta
	var direrr error

	wg.Add(1)

	go func() {
		defer wg.Done()
		dir, direrr = d.client.GetDir(path, "", "", 0, 0, "", "", token.AccessToken)
	}()

	meta, err := d.client.GetMeta(path, "", "", token.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	if meta.Filetype == "dir" {
		wg.Wait()
		if direrr != nil {
			return nil, errors.Wrap(direrr, "")
		}
		meta.Members = dir.Members
	}

	return meta, nil
}

func (d *Drive) Listdir(path string, authkey string) (*Meta, error) {

	token, err := d.manager.GetAuthToken(authkey)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t get valid auth token for authkey %q", authkey)
	}
	memberfields := "members.id,members.name,members.size,members.nmembers,members.type,members.mime_type,members.ctime,members.mtime,members.image.height,members.image.width,members.readable,members.writable"

	params := url.Values{
		"path":    {d.clean(path, token.Alias)},
		"members": {"all"},
		"fields":  {metafields + "," + memberfields},
	}
	// fmt.Println("params", params, token)
	body, err := d.client.Request("GET", "/dir", params, nil, token.AccessToken)
	if err != nil {
		fmt.Println("error", err)
		return nil, errors.Wrap(err)
	}
	defer body.Close()
	// b, err := io.ReadAll(body)
	// fmt.Println("params", string(b), err)
	var response Meta
	err = json.NewDecoder(body).Decode(&response)
	if err != nil {
		// fmt.Println("params", params, token)
		return nil, errors.Wrap(err)
	}
	return &response, nil
}

func (d *Drive) Mkdir(path string, authkey string) (*Meta, error) {

	token, err := d.manager.GetAuthToken(authkey)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t get valid auth token for authkey %q", authkey)
	}
	_, err = d.client.PostDir(d.clean(path, token.Alias), "", "", 0, 0, token.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "Error in post request")
	}

	return nil, nil
}

func (d *Drive) Rmdir(dirname, authkey string) error {

	token, _ := d.manager.GetAuthToken(authkey)
	params := url.Values{
		"path":      {d.clean(dirname, token.Alias)},
		"recursive": {"true"},
	}

	_, err := d.client.Request("DELETE", "/dir", params, nil, token.AccessToken)
	if err != nil {
		return errors.Wrap(err, "Error in delete request")
	}
	// defer readcloser.Close()

	return nil
}

func (d *Drive) Rm(filename string, authkey string) error {

	token, _ := d.manager.GetAuthToken(authkey)
	params := url.Values{
		"path": {d.clean(filename, token.Alias)},
	}

	_, err := d.client.Request("DELETE", "/file", params, nil, token.AccessToken)
	if err != nil {
		return errors.Wrap(err, "Error in delete request")
	}

	return nil
}
func (d *Drive) CreateFile(path string, body io.Reader, name string, modtime string, authuser string) (*Meta, error) {

	token, err := d.manager.GetAuthToken(authuser)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t get valid auth token for authuser %q", authuser)
	}
	params := url.Values{
		"dir":      {d.clean(path, token.Alias)},
		"name":     {name},
		"on_exist": {"autoname"},
	}
	if modtime != "" {
		params.Set("mtime", modtime)
	}
	respBody, err := d.client.Request("POST", "/file", params, body, token.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "Error in post request")
	}
	defer respBody.Close()

	bytes, err := ioutil.ReadAll(respBody)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading request body")
	}
	fmt.Println("createfile:", err, path, name, modtime, authuser, token, string(bytes))

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

	meta := Meta{r.ID, r.Name, r.Path, r.Type, r.CTime, r.MTime, r.HasDirs, r.Readable, r.Writable, r.Size, r.MIMEType, 0, "", nil, nil}
	if r.Image != nil {
		meta.Image = &Image{Height: r.Image.Height, Width: r.Image.Width}
	}

	return &meta, nil
}

func (d *Drive) Save(filepath string, body io.Reader, authuser string) (*Meta, error) {

	token, err := d.manager.GetAuthToken(authuser)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t get valid auth token for authuser %q", authuser)
	}

	dir, file := path.Split(d.clean(filepath, token.Alias))
	dir = strings.TrimSuffix(dir, "/")
	meta, err := d.client.PutFile(body, dir, file, 0, 0, token.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "Error in hidrive PUT file request")
	}

	return meta, nil
}
