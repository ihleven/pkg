package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/ihleven/cloud11-api/auth"
	"github.com/ihleven/cloud11-api/drive"
)

var Drive FSWebDrive = FSWebDrive{Root: "/Users/mi/data/tmp", Prefix: "/home", ServeURL: "/serve/home", PermissionMode: 0}

type FSWebDrive struct {
	// Absolute path inside filesystem
	Root string `json:"-"` // /home/ihle/tmp
	// pathname of root dir in webview
	Prefix string `json:"url"` // /home
	// pathname of root dir in serveview
	ServeURL string `json:"serveUrl"` // /serve/home
	// indicates if index.html is served for directories in serveview
	serveIndexHtml bool `json:"-"`
	// AlbumURL       string          `json:"albumUrl"`

	// Owner          *drive.User     `json:"-"` // alle Dateien gehören automatisch diesem User ( => homes )
	// Group          *drive.Group    `json:"-"` // jedes File des Storage bekommt automatisch diese Gruppe ( z.B. brunhilde )

	// PermMode overwrites file permissions globally (if set).
	// e.g., for a public readable but not writable filesystem: 0444
	PermissionMode os.FileMode              `json:"-"` // wenn gesetzt erhält jedes File dies Permission =< wird nicht mehr auf fs gelesen
	Accounts       map[string]*auth.Account `json:"-"`
}

func (wd *FSWebDrive) GetHomeHandle(name string, user *auth.Account) (drive.Handle, error) {

	if user.HomeDir == "" {
		return nil, errors.New("Authemtication required")
	}
	location := path.Join(user.HomeDir, name)

	info, err := os.Stat(location)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "os.Stat failed for %s (location: %s)", name, location)
		}
		return nil, errors.Wrapf(err, "os.Stat failed for %s (location: %s)", name, location)
	}

	fh := handle{FileInfo: info, location: location, mode: info.Mode()}

	if wd.PermissionMode != 0 {
		// replace 9 least significant bits from mode with storage.PermissionMode
		fh.mode = (fh.mode & 0xfffffe00) | (wd.PermissionMode & os.ModePerm) // & 0x1ff
	}
	return &fh, nil
}

func (wd *FSWebDrive) GetServeHandle(path string) (os.FileInfo, *os.File, error) {

	location := filepath.Join(wd.Root, path)
	//location := strings.Replace(filepath.Clean(path), wd.ServeURL, wd.Root, 1)
	fmt.Println("GetServeHandle", path, location)

	info, err := os.Stat(location)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, errors.Wrapf(err, "os.Stat failed for %s (location: %s)", path, location)
		}
		return nil, nil, errors.Wrapf(err, "os.Stat failed for %s (location: %s)", path, location)
	}

	fd, err := os.OpenFile(location, os.O_RDONLY, 0)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "OpenFile failed for %s (location: %s)", path, location)
	}

	return info, fd, nil
}

func (wd *FSWebDrive) Open(name string) (drive.Handle, error) {
	// path := strings.TrimPrefix(filepath.Clean(url), prefix)
	// location := filepath.Join(wd.Root, path)
	location := path.Join(wd.Root, name)

	info, err := os.Stat(location)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "os.Stat failed for %s (location: %s)", name, location)
		}
		return nil, errors.Wrapf(err, "os.Stat failed for %s (location: %s)", name, location)
	}

	fh := handle{FileInfo: info, location: location, mode: info.Mode()}

	if wd.PermissionMode != 0 {
		// replace 9 least significant bits from mode with storage.PermissionMode
		fh.mode = (fh.mode & 0xfffffe00) | (wd.PermissionMode & os.ModePerm) // & 0x1ff
	}
	return &fh, nil
}
func (wd *FSWebDrive) OpenFile(name string, account *auth.Account) (*drive.File, error) {
	return wd.GetFile(name, account)
}
func (wd *FSWebDrive) GetHandle(name string, t drive.PathType) (drive.Handle, error) {
	// path := strings.TrimPrefix(filepath.Clean(url), prefix)
	// location := filepath.Join(wd.Root, path)
	location := path.Join(wd.Root, name)
	if t == drive.AbsPath {
		location = name
	}

	info, err := os.Stat(location)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "os.Stat failed for %s (location: %s)", name, location)
		}
		return nil, errors.Wrapf(err, "os.Stat failed for %s (location: %s)", name, location)
	}

	fh := handle{FileInfo: info, location: location, mode: info.Mode()}

	if wd.PermissionMode != 0 {
		// replace 9 least significant bits from mode with storage.PermissionMode
		fh.mode = (fh.mode & 0xfffffe00) | (wd.PermissionMode & os.ModePerm) // & 0x1ff
	}
	return &fh, nil
}

func (wd *FSWebDrive) GetFile(name string, account *auth.Account) (*drive.File, error) {

	Handle, err := wd.GetHomeHandle(name, account)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not GetHandle(%v)", name)
	}
	h := Handle.(*handle)

	uid, gid := h.userAndGroupIDs()
	var file = drive.File{
		Handle:        h,
		URL:           path.Join(wd.Prefix, name),
		Name:          h.Name(),
		Size:          h.Size(),
		Mode:          h.mode,
		Type:          h.GuessMIME(),
		Permissions:   h.mode.String(),
		Owner:         GetUserByID(uid),
		Group:         GetGroupByID(gid),
		Authorization: h.GetPermissions(account),
		Modified:      h.ModTime(),
	}

	return &file, nil
}

func (wd *FSWebDrive) ListFiles(folder *drive.File, account *auth.Account) ([]drive.File, error) {

	handles, err := folder.ReadDir(wd.PermissionMode)
	if err != nil {
		return nil, err
	}

	entries := make([]drive.File, len(handles))

	for index, Handle := range handles {
		h := Handle.(*handle)
		uid, gid := h.userAndGroupIDs()
		entries[index] = drive.File{
			Handle:        h,
			URL:           filepath.Join(folder.URL, h.Name()),
			Name:          h.Name(),
			Size:          h.Size(),
			Mode:          h.mode,
			Type:          h.GuessMIME(),
			Permissions:   h.mode.String(),
			Owner:         GetUserByID(uid),
			Group:         GetGroupByID(gid),
			Authorization: h.GetPermissions(account),
			Modified:      h.ModTime(),
		}
	}
	return entries, nil
}

func (wd *FSWebDrive) CreateFile(folder *drive.File, name string) (drive.Handle, error) {
	h := folder.Handle.(*handle)
	l := path.Join(h.location, name)
	//
	var _, err = os.Stat(l)
	var file *os.File
	// create file if not exists
	if os.IsNotExist(err) {
		file, err = os.Create(l)

		//defer file.Close()
	} else {
		basename := strings.TrimSuffix(name, filepath.Ext(name)) + ".*" + filepath.Ext(name)
		file, err = ioutil.TempFile(h.location, basename)
	}
	// file, err := os.OpenFile(h.location+"/"+name, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not create file %v", name)
	}
	info, err := file.Stat()
	handle := NewHandle(info, path.Join(h.location, info.Name()), 0)
	return handle, nil
}

func (wd *FSWebDrive) location(p string, pathtype drive.PathType) string {
	location := p
	switch pathtype {
	case drive.URLPath:
		location = strings.Replace(p, wd.Prefix, wd.Root, 1)
	case drive.AbsPath:
	default:
		location = path.Join(wd.Root, p)
	}
	return location
}

// Create creates an empty file
func (wd *FSWebDrive) Create(url string, pathtype drive.PathType) (drive.Handle, error) {

	location := wd.location(url, pathtype)
	fmt.Println("Create: ", location)
	var _, err = os.Stat(location)

	var file *os.File
	dir, base := path.Split(location)

	if os.IsNotExist(err) {
		file, err = os.Create(location)
		defer file.Close()
	} else {
		basename := strings.TrimSuffix(base, filepath.Ext(base)) + ".*" + filepath.Ext(base)
		file, err = ioutil.TempFile(dir, basename)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "Could not create file %v", url)
	}

	info, err := file.Stat()
	if err != nil {
		return nil, errors.Wrapf(err, "Could not stat/create file %v", url)
	}

	return NewHandle(info, path.Join(dir, info.Name()), 0), nil
}

func (wd *FSWebDrive) Mkdir(url string) (drive.Handle, error) {

	location := strings.Replace(path.Clean(url), wd.Prefix, wd.Root, 1)

	err := os.Mkdir(location, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not create folder %v", url)
	}
	handle, err := wd.GetHandle(path.Join(wd.Root, url), drive.AbsPath)

	return handle, nil
}
