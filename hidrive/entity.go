package hidrive

import (
	"io/fs"
	"net/url"
	"time"
)

// meta
// left out: ,has_dirs,parent_id, rshare,shareable,teamfolder,zone     "chash,mhash,mohash,nhash"
var metafields = "id,name,path,type,mime_type,ctime,mtime,readable,writable,size,nmembers"
var imagefields = "image.exif,image.width,image.height"
var memberfields = "members.name,members.readable,members.writable,members.type,members.mime_type,members.mtime,members.ctime,members.size,members.nmembers,members.id,members.image.height,members.image.width,members.image.exif"

type Meta struct {
	NameURLEncoded string `json:"name"`           //    - string    - URL-encoded name of the directory
	Path           string `json:"path,omitempty"` //    - string    - URL-encoded path to the directory
	Filetype       string `json:"type"`           //    - string    - e.g. "dir"
	MIMEType       string `json:"mime_type"`
	MTime          int64  `json:"mtime,omitempty"`    //    - timestamp - mtime of the directory
	CTime          int64  `json:"ctime,omitempty"`    //    - timestamp - ctime of the directory
	Readable       bool   `json:"readable,omitempty"` //    - bool      - read-permission for the directory
	Writable       bool   `json:"writable,omitempty"` //    - bool      - write-permission for the directory
	Filesize       uint64 `json:"size,omitempty"`
	Nmembers       int    `json:"nmembers,omitempty"`
	HasDirs        bool   `json:"has_dirs,omitempty"` //    - bool      - does the directory contain subdirs?
	ID             string `json:"id,omitempty"`       //    - string    - path id of the directory
	ParentID       string `json:"parent_id,omitempty"`
	// default: ctime,has_dirs,mtime,readable,size,type,writable
	// Chash    string      `json:"chash"`
	// Mhash    string      `json:"mhash"`
	// MOhash   string      `json:"mohash"`
	// Nhash    string      `json:"nhash"`
	Image   *Image `json:"image,omitempty"`
	Members []Meta `json:"members,omitempty"`
	Content string `json:"content,omitempty"`
}

type Image struct {
	Height int  `json:"height"`
	Width  int  `json:"width"`
	Exif   Exif `json:"exif"`
}

type Exif struct {
	// 	Aperture         string
	// 	BitsPerSample    string
	DateTimeOriginal string
	ExifImageHeight  interface{} // string
	ExifImageWidth   interface{} // string
	// 	ExposureTime     string
	// 	FocalLength      string
	// 	ISO              string
	// 	ImageHeight      string
	// 	ImageWidth       string
	Make        string
	Model       string
	Orientation interface{} // string
	// 	ResolutionUnit   string
	// 	XResolution      string
	// 	YResolution      string
}

// Name is part of fs.FileInfo and fs.DirEntry interface
func (m *Meta) Name() string { // base name of the file
	unescapedName, _ := url.QueryUnescape(m.NameURLEncoded)
	return unescapedName
}

// IsDir is part of fs.FileInfo and fs.DirEntry interface
func (m *Meta) IsDir() bool { // abbreviation for Mode().IsDir()
	return m.Filetype == "dir"
}

// Type is part of fs.DirEntry interface
func (e *Meta) Type() fs.FileMode {
	// TODO!!!
	var mode uint32
	return fs.FileMode(mode)
}

// Info is part of fs.DirEntry interface
func (e *Meta) Info() (fs.FileInfo, error) {
	return e, nil
}

// Size is part of fs.FileInfo interface
func (m *Meta) Size() int64 { // length in bytes for regular files; system-dependent for others
	return int64(m.Filesize)
}

// Mode is part of fs.FileInfo interface
func (m *Meta) Mode() fs.FileMode { // file mode bits
	var mode uint32
	return fs.FileMode(mode)
}

// ModTime is part of fs.FileInfo interface
func (m *Meta) ModTime() time.Time {
	return time.Unix(0, m.MTime)
}

// Sys is part of fs.FileInfo interface
func (m *Meta) Sys() interface{} {
	return m
}
