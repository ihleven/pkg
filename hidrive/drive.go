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
func NewDrive(prefix string, tokenmngr *AuthManager) *Drive {

	var Drive = Drive{
		client:  NewClient(),
		manager: tokenmngr,
		prefix:  prefix,
		// homes:    map[string]string{"matt": "/users/matt.ihle"},
		confmap: make(map[string]config),
	}
	return &Drive
}

type Drive struct {
	client  *HiDriveClient
	manager *AuthManager // Auth     *OAuth2Prov
	prefix  string
	// homes    map[string]string
	confmap map[string]config
}

type config struct {
	Username string
	Prefix   string
	ACL      []string
}

func (d *Drive) clean(inpath string, username string) string {

	outpath := path.Clean(inpath)

	if d.prefix != "" {
		outpath = path.Join(d.prefix, outpath)
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
func (d *Drive) GetMeta(path string, authkey string) (interface{}, error) {
	fmt.Println("GetMeta:", path, authkey)
	token, err := d.manager.GetAuthToken(authkey)
	if token == nil {
		return nil, errors.NewWithCode(401, "no valid token")
	}
	fmt.Println("token:", token)
	path = d.clean(path, token.Alias)
	fmt.Println("path:", path)
	var wg sync.WaitGroup
	var dir *Dir
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
	fmt.Println("meta:", meta)
	if meta.Base.Type == "dir" {
		wg.Wait()
		if direrr != nil {
			return nil, errors.Wrap(direrr, "")
		}
		return dir, nil
	}

	return &meta, nil

	// switch meta.Type {
	// case "dir":
	// 	dir, err := d.GetDir(pfad, authuser)
	// 	return dir, nil
	// }

}
func (d *Drive) GetDir(path string, authkey string) (*Dir, error) {

	params := make(map[string][]string)

	memberfields := "members,members.id,members.name,members.nmembers,members.size,members.type,members.mime_type,members.mtime,members.image.height,members.image.width,members.image.exif"
	params["path"] = []string{d.clean(path, authkey)}
	params["members"] = []string{"all"}
	params["fields"] = []string{metafields + "," + memberfields}

	token, _ := d.manager.GetAuthToken(authkey)
	body, err := d.client.Request("GET", "/dir", params, nil, token.AccessToken)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var response Dir
	err = json.NewDecoder(body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (d *Drive) Mkdir(path string, authkey string) (*Meta, error) {

	token, _ := d.manager.GetAuthToken(authkey)
	_, err := d.client.PostDir(d.clean(path, authkey), "", "", 0, 0, token.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "Error in post request")
	}

	return nil, nil
}

func (d *Drive) Rmdir(dirname, authkey string) error {

	params := url.Values{
		"path":      {d.clean(dirname, authkey)},
		"recursive": {"true"},
	}
	token, _ := d.manager.GetAuthToken(authkey)
	readcloser, err := d.client.Request("DELETE", "/dir", params, nil, token.AccessToken)
	if err != nil {
		return errors.Wrap(err, "Error in post request")
	}
	defer readcloser.Close()

	return nil
}

func (d *Drive) Rm(filename string, authkey string) error {

	params := url.Values{
		"path": {d.clean(filename, authkey)},
	}
	token, _ := d.manager.GetAuthToken(authkey)
	_, err := d.client.Request("DELETE", "/file", params, nil, token.AccessToken)
	if err != nil {
		return errors.Wrap(err, "Error in delete request")
	}

	return nil
}
func (d *Drive) CreateFile(path string, body io.Reader, name string, modtime string, authuser string) (*Meta, error) {

	token, err := d.manager.GetAccessToken(authuser)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t get valid auth token for authuser %q", authuser)
	}
	respBody, err := d.client.Request("POST", "/file", url.Values{
		"dir":      {path},
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

	meta := Meta{Base{r.ID, r.Name, r.Path, r.Type, r.CTime, r.MTime, r.HasDirs, r.Readable, r.Writable}, r.MIMEType, r.Size, 0, "", nil}
	if r.Image != nil {
		meta.Image = &Image{Height: r.Image.Height, Width: r.Image.Width}
	}

	return &meta, nil
}
