package hidrive

import (
	"io/fs"
	"net/url"
	"time"
)

var defaultMetaFields = "ctime,has_dirs,mtime,readable,size,type,writable"

var allDirFields = "chash,ctime,has_dirs,id,members,members.chash,members.ctime,members.has_dirs,members.id,members.image.exif,members.image.height,members.image.width,members.mhash,members.mime_type,members.mohash,members.mtime,members.name,members.nmembers,members.nhash,members.parent_id,members.path,members.readable,members.rshare,members.size,members.type,members.writable,mhash,mohash,mtime,name,nhash,nmembers,parent_id,path,readable,rshare,size,type,writable"

var dirfields = "id,name,path,type,mime_type,ctime,mtime,readable,writable,size,nmembers,has_dirs,parent_id,rshare,shareable,teamfolder"
var metafields = "id,name,path,type,mime_type,ctime,mtime,readable,writable,size,nmembers,has_dirs,parent_id"
var extrafields = "rshare,shareable,teamfolder,zone"
var imagefields = "image.exif,image.width,image.height"
var defaultfields = "ctime,has_dirs,mtime,readable,size,type,writable"
var memberfields = "members,members.ctime,members.has_dirs,members.id,members.image.exif,members.image.height,members.image.width,members.mime_type,members.mtime,members.name,members.nmembers,members.parent_id,members.path,members.readable,members.rshare,members.size,members.type,members.writable"

type Meta struct {
	ID             string `json:"id"`       //    - string    - path id of the directory
	NameURLEncoded string `json:"name"`     //    - string    - URL-encoded name of the directory
	Path           string `json:"path"`     //    - string    - URL-encoded path to the directory
	Filetype       string `json:"type"`     //    - string    - e.g. "dir"
	CTime          int64  `json:"ctime"`    //    - timestamp - ctime of the directory
	MTime          int64  `json:"mtime"`    //    - timestamp - mtime of the directory
	HasDirs        bool   `json:"has_dirs"` //    - bool      - does the directory contain subdirs?
	Readable       bool   `json:"readable"` //    - bool      - read-permission for the directory
	Writable       bool   `json:"writable"` //    - bool      - write-permission for the directory
	// default: ctime,has_dirs,mtime,readable,size,type,writable

	Filesize uint64 `json:"size"`
	MIMEType string `json:"mime_type"`
	Nmembers int    `json:"nmembers"`
	ParentID string `json:"parent_id"`
	// Chash    string      `json:"chash"`
	// Mhash    string      `json:"mhash"`
	// MOhash   string      `json:"mohash"`
	// Nhash    string      `json:"nhash"`
	Image   *Image `json:"image"`
	Members []Meta `json:"members"`
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
	ExifImageHeight  int // string
	ExifImageWidth   int // string
	// 	ExposureTime     string
	// 	FocalLength      string
	// 	ISO              string
	// 	ImageHeight      string
	// 	ImageWidth       string
	// 	Make             string
	// 	Model            string
	Orientation int // string
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
