package hidrive

import (
	"encoding/json"
	"io"
	"path"
	"strings"

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

func (d *Drive) fullpath(drivepath string, username string) string {

	cleanpath := path.Clean(drivepath)

	if d.prefix != "" {
		return path.Join(d.prefix, cleanpath)
	} else if d.useHome {
		return path.Join("/users", username, cleanpath)
	}
	// else if username != "" {
	// 	if strings.HasPrefix(outpath, "/home") {
	// 		outpath = strings.Replace(outpath, "/home", "/users/"+username, 1)
	// 	}
	// 	if strings.HasPrefix(outpath, "/~") {
	// 		outpath = strings.Replace(outpath, "/~", "/users/"+username, 1)
	// 	}
	// }
	// else {
	//     tail = strings.Replace(tail, "/home", "/", 1)
	// }
	return cleanpath
}

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

func (d *Drive) parse(response io.Reader) (*Meta, error) {
	var meta Meta
	err := json.NewDecoder(response).Decode(&meta)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding response")
	}
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

func (d *Drive) Tkn(username string) *AuthToken {

	token, err := d.manager.GetAuthToken(username)
	if err != nil {
		return nil
	}
	return token
}
