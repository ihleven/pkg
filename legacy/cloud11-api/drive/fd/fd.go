package fd

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

type fd interface {
	io.ReadCloseWriter
	io.Seeker
}

func ServeDescriptor(wd Driver) func(w http.ResponseWriter, r *http.Request) {

	var dispatchRaw http.HandlerFunc

	dispatchRaw = func(w http.ResponseWriter, r *http.Request) {

		//authuser, err := session.GetSessionUser(r, w)

		fd, err := wd.GetFD(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer fd.Close()

		if fd.IsDir() {
			r.URL.Path = path.Join(r.URL.Path, "index.html")
			Serve(webdrive)(w, r)
			return
		}

		http.ServeContent(w, r, fd.name, fd.mtime, fd)
	}
	return dispatchRaw
}

type descriptor struct {
	*os.File
	stat     os.FileInfo
	location string
	name     string
	//size
	mode  os.FileMode
	mtime *time.Time
	//atime
	//ctime
	//isdir
}

func (wd *FSWebDrive) GetFD(name string) (fd, error) {

	fd := descriptor{
		location: path.Join(wd.Root, filepath.Clean(name)),
	}

	f.File, err = os.OpenFile(f.location, os.O_RDONLY, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "os.FileOpen failed for %s (location: %s)", name, location)
	}

	f.stat, err = f.File.Stat()
	if err != nil {
		return nil, errors.Wrapf(err, "os.Stat failed for %s (location: %s)", path, location)
	}
	f.name = f.stat.Name()

	return info, fd, nil
}
