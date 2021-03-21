package hidrive

import (
	"bytes"
	"io"
	"io/fs"
	"net/url"
	"path"

	"github.com/ihleven/errors"
)

func NewFS(prefix, token string) *FS {
	return &FS{
		client: NewClient(),
		prefix: prefix,
		token:  token,
	}
}

type FS struct {
	client *HiDriveClient
	prefix string
	token  string // *Token2
}

func (fs *FS) Open(name string) (fs.File, error) {

	meta, err := fs.client.GetMeta(path.Join(fs.prefix, name), "", "", fs.token)
	if err != nil {
		return nil, errors.Wrap(err, "Error opening file %q", name)
	}
	return &File{meta: meta, fsys: fs}, nil
}

func (fsys *FS) ReadDir(name string) ([]fs.DirEntry, error) {

	dir, err := fsys.client.GetDir(path.Join(fsys.prefix, name), "", "", 0, 0, "", "", fsys.token)
	if err != nil {
		return nil, err
	}

	// https://stackoverflow.com/a/12994852
	var entries = make([]fs.DirEntry, len(dir.Members))
	for i := range dir.Members {
		entries[i] = &dir.Members[i]
	}
	return entries, nil
}

func (fs *FS) ReadFile(name string) ([]byte, error) {

	params := url.Values{"path": {path.Join(fs.prefix, name)}}

	body, err := fs.client.Request("GET", "/file", params, nil, fs.token)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	defer body.Close()
	bytes, err := io.ReadAll(body)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return bytes, nil
}

type File struct {
	meta   *Meta
	fsys   *FS
	reader *bytes.Reader
}

func (f *File) Stat() (fs.FileInfo, error) {
	return f.meta, nil
}

func (f *File) read() error {
	body, err := f.fsys.client.GetFile(f.meta.Path, "", f.fsys.token)
	if err != nil {
		return err
	}
	defer body.Close()

	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	f.reader = bytes.NewReader(b)
	return nil
}
func (f *File) Read(buf []byte) (int, error) {

	if f.reader == nil {
		if err := f.read(); err != nil {
			return 0, err
		}
	}
	return f.reader.Read(buf)
}

func (f *File) Seek(offset int64, whence int) (int64, error) {

	if f.reader == nil {
		if err := f.read(); err != nil {
			return 0, err
		}
	}
	return f.reader.Seek(offset, whence)
}

func (f *File) Close() error {
	f.reader = nil
	return nil
}
