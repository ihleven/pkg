package hidrive

import (
	"io/fs"
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

// type DirResponse struct {
// 	Meta
// 	Members []Meta `json:"members"`
// }

// func (d *DirResponse) Respond(w http.ResponseWriter, contenttype string) {
// 	encoder := json.NewEncoder(w)
// 	encoder.SetIndent("", "    ")
// 	err := encoder.Encode(d)
// 	if err != nil {
// 		http.Error(w, err.Error(), 500)
// 	}
// }

// type Dir struct {
// 	ID       string    //    - string    - path id of the directory
// 	Name     string    //    - string    - URL-encoded name of the directory
// 	Path     string    //    - string    - URL-encoded path to the directory
// 	Type     string    //    - string    - "dir"
// 	Ctime    string    //    - timestamp - ctime of the directory
// 	Mtime    time.Time //    - timestamp - mtime of the directory
// 	HasDirs  bool      //    - bool      - does the directory contain subdirs?
// 	Readable bool      //    - bool      - read-permission for the directory
// 	Writable bool      //    - bool      - write-permission for the directory
// }

type Base struct {
	ID       string //    - string    - path id of the directory
	Name     string //    - string    - URL-encoded name of the directory
	Path     string //    - string    - URL-encoded path to the directory
	Type     string //    - string    - "dir"
	Ctime    int64  //    - timestamp - ctime of the directory
	Mtime    int64  // time.Time //    - timestamp - mtime of the directory
	HasDirs  bool   //    - bool      - does the directory contain subdirs?
	Readable bool   //    - bool      - read-permission for the directory
	Writable bool   //    - bool      - write-permission for the directory
}

type Meta struct {
	Base
	// ID       string `json:"id"`
	// Name     string `json:"name"`
	// Path     string `json:"path"`
	// Type     string `json:"type"`
	// CTime    int64  `json:"ctime"`
	// MTime    int64  `json:"mtime"`
	// HasDirs  bool   `json:"has_dirs"`
	// Readable bool   `json:"readable"`
	// Writable bool   `json:"writable"`
	// default: ctime,has_dirs,mtime,readable,size,type,writable

	MIMEType string `json:"mime_type"`
	Filesize uint64 `json:"size"`
	Nmembers int    `json:"nmembers"`
	ParentID string `json:"parent_id"`
	// Chash    string      `json:"chash"`
	// Mhash    string      `json:"mhash"`
	// MOhash   string      `json:"mohash"`
	// Nhash    string      `json:"nhash"`
	Image *Image `json:"image"`
}
type Dir struct {
	Meta
	Members []Meta `json:"members"`
}

// Name is part of fs.FileInfo and fs.DirEntry interface
func (m *Meta) Name() string { // base name of the file
	return m.Base.Name
}

// IsDir is part of fs.FileInfo and fs.DirEntry interface
func (m *Meta) IsDir() bool { // abbreviation for Mode().IsDir()
	return m.Base.Type == "dir"
}

// Type is part of fs.DirEntry interface
func (e *Meta) Type() fs.FileMode {
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
	return time.Unix(0, m.Base.Mtime)
}

// Sys is part of fs.FileInfo interface
func (m *Meta) Sys() interface{} {
	return m.Base
}

// type FileDirSymlinkMeta struct {
// 	ID         string `json:"id"`
// 	NameHidden string `json:"name"`
// 	Path       string `json:"path"`
// 	Type       string `json:"type"`
// 	CTime      int64  `json:"ctime"`
// 	MTime      int64  `json:"mtime"`
// 	HasDirs    bool   `json:"has_dirs"`
// 	Readable   bool   `json:"readable"`
// 	Writable   bool   `json:"writable"`
// 	SizeVar    uint64 `json:"size"`
// 	// MIMEType string `json:"mime_type"`
// 	// ParentID string `json:"parent_id"`
// 	// Nmembers int    `json:"nmembers"`
// 	// reader *bytes.Reader
// }

// func (m *FileDirSymlinkMeta) Stat() (fs.FileInfo, error) {
// 	return m, nil
// }

// func (m *FileDirSymlinkMeta) Read(buf []byte) (int, error) {
// 	if m.reader == nil {
// 		buffer := make([]byte, m.SizeVar)
// 		m.reader = bytes.NewReader(buffer)
// 	}

// 	return m.reader.Read(buf)
// }
// func (m *FileDirSymlinkMeta) Seek(offset int64, whence int) (int64, error) {
// 	if m.reader == nil {
// 		buffer := make([]byte, m.SizeVar)
// 		m.reader = bytes.NewReader(buffer)
// 	}
// 	return m.reader.Seek(offset, whence)
// }

// func (m *FileDirSymlinkMeta) Close() error {
// 	return nil
// }

// func (m *Meta) Respond(w http.ResponseWriter, contenttype string) {
// 	encoder := json.NewEncoder(w)
// 	encoder.SetIndent("", "    ")
// 	err := encoder.Encode(m)
// 	if err != nil {
// 		http.Error(w, err.Error(), 500)
// 	}
// }

type Image struct {
	Height int  `json:"height"`
	Width  int  `json:"width"`
	Exif   Exif `json:"exif"`
}

type Exif struct {
	DateTimeOriginal string
	ExifImageHeight  int
	ExifImageWidth   int

	Orientation int
}

// type Exif2 struct {
// 	Aperture         string
// 	BitsPerSample    string
// 	DateTimeOriginal string
// 	ExifImageHeight  string
// 	ExifImageWidth   string
// 	ExposureTime     string
// 	FocalLength      string
// 	ISO              string
// 	ImageHeight      string
// 	ImageWidth       string
// 	Make             string
// 	Model            string
// 	Orientation      string
// 	ResolutionUnit   string
// 	XResolution      string
// 	YResolution      string
// }

// type Responder interface {
// 	Respond(http.ResponseWriter, string)
// }
