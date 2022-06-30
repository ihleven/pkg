package hidrive

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/ihleven/errors"
)

// NewDrive creates a new hidrive
func NewDrive(manager *AuthManager, opts ...DriveOption) *Drive {

	var d = &Drive{
		client:  NewClient(),
		manager: manager,
		// prefix:  "",
		// homes:    map[string]string{"matt": "/users/matt.ihle"},
		confmap: make(map[string]config),
	}

	// Loop through and apply each option
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

// translate path relative to d to full hidrive path
func (d *Drive) fullpath(drivepath string, username string) string {

	cleanpath := path.Clean(drivepath)

	if d.prefix != "" {
		return path.Join(d.prefix, cleanpath)
	} else if d.useHome {
		return path.Join("/users", username, cleanpath)
	}

	return cleanpath
}

// translate absolute hidrive path to relative drive path
func (d *Drive) drivepath(fullpath string, username string) string {

	if d.prefix != "" {
		return strings.TrimPrefix(fullpath, d.prefix)
	} else if d.useHome {
		startsWithUsername := strings.TrimPrefix(fullpath, "/users/")
		pathWithoutSlash := strings.SplitAfterN(startsWithUsername, "/", 2)
		if len(pathWithoutSlash) > 1 {
			return "/" + pathWithoutSlash[1]
		} else {
			return "/"
		}
	}

	return fullpath
}

func (d *Drive) processMetaResponse(response io.Reader) (*Meta, error) {
	// fmt.Println("processMetaResponse:")
	var meta Meta
	err := json.NewDecoder(response).Decode(&meta)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding response")
	}
	// fmt.Println("processMetaResponse:", meta)
	meta.Path = d.drivepath(meta.Path, "")
	meta.NameURLEncoded = meta.Name()

	if meta.Path == "" {
		meta.Path = "/"
	}
	if meta.Path == "/" {
		meta.NameURLEncoded = ""
	}
	return &meta, nil
}

func (d *Drive) processMetaResponseCreateFile(response io.Reader) (*Meta, error) {
	type Exif2 struct {
		DateTimeOriginal string
		ExifImageHeight  string
		ExifImageWidth   string
	}
	type Image2 struct {
		Height int   `json:"height"`
		Width  int   `json:"width"`
		Exif   Exif2 `json:"exif"`
	}
	type Metana struct {
		NameURLEncoded string  `json:"name"`           //    - string    - URL-encoded name of the directory
		Path           string  `json:"path,omitempty"` //    - string    - URL-encoded path to the directory
		Filetype       string  `json:"type"`           //    - string    - e.g. "dir"
		MIMEType       string  `json:"mime_type"`
		MTime          int64   `json:"mtime,omitempty"`    //    - timestamp - mtime of the directory
		CTime          int64   `json:"ctime,omitempty"`    //    - timestamp - ctime of the directory
		Readable       bool    `json:"readable,omitempty"` //    - bool      - read-permission for the directory
		Writable       bool    `json:"writable,omitempty"` //    - bool      - write-permission for the directory
		Filesize       uint64  `json:"size,omitempty"`
		Nmembers       int     `json:"nmembers,omitempty"`
		HasDirs        bool    `json:"has_dirs,omitempty"` //    - bool      - does the directory contain subdirs?
		ID             string  `json:"id,omitempty"`       //    - string    - path id of the directory
		ParentID       string  `json:"parent_id,omitempty"`
		Image          *Image2 `json:"image,omitempty"`
		Members        []Meta  `json:"members,omitempty"`
		Content        string  `json:"content,omitempty"`
	}

	var m Metana

	err := json.NewDecoder(response).Decode(&m)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding response")
	}
	fmt.Println("metana: ", m)
	meta := Meta{
		NameURLEncoded: m.NameURLEncoded,
		Path:           d.drivepath(m.Path, ""),
		Image: &Image{
			Width:  m.Image.Width,
			Height: m.Image.Height,
			Exif:   Exif{DateTimeOriginal: m.Image.Exif.DateTimeOriginal},
		},
	}

	if meta.Path == "" {
		meta.Path = "/"
	}
	if meta.Path == "/" {
		meta.NameURLEncoded = ""
	}
	return &meta, nil
}

// func (d *Drive) Tkn(username string) *AuthToken {

// 	token, err := d.manager.GetAccessToken(username)
// 	if err != nil {
// 		return nil
// 	}
// 	return token
// }

// func (d *Drive) credentialsDeprecated(r *http.Request) (string, string) {

// 	claims, _, err := auth.GetClaims(r)
// 	if err != nil {
// 		fmt.Println("MauthManager.credentials -1- ERROR:", err)
// 		return "", ""
// 	}

// 	token, err := d.manager.GetAccessToken(claims.Username)
// 	if err != nil {
// 		fmt.Println("MauthManager.credentials - 2 - ERROR:", err)
// 		return "", ""
// 	}
// 	return d.fullpath(r.URL.Path, token.Alias), token.AccessToken
// }

func (d *Drive) Token(authkey string) *Token {
	token, err := d.manager.GetAccessToken(authkey)
	if err != nil {
		return nil
	}
	return token
}
