package hidrive

import (
	"encoding/json"
	"net/http"
)

var defaultMetaFields = "ctime,has_dirs,mtime,readable,size,type,writable"

var allDirFields = "chash,ctime,has_dirs,id,members,members.chash,members.ctime,members.has_dirs,members.id,members.image.exif,members.image.height,members.image.width,members.mhash,members.mime_type,members.mohash,members.mtime,members.name,members.nmembers,members.nhash,members.parent_id,members.path,members.readable,members.rshare,members.size,members.type,members.writable,mhash,mohash,mtime,name,nhash,nmembers,parent_id,path,readable,rshare,size,type,writable"

var dirfields = "id,name,path,type,mime_type,ctime,mtime,readable,writable,size,nmembers,has_dirs,parent_id,rshare,shareable,teamfolder"
var metafields = "id,name,path,type,mime_type,ctime,mtime,readable,writable,size,nmembers,has_dirs,parent_id"
var extrafields = "rshare,shareable,teamfolder,zone"
var imagefields = "image.exif,image.width,image.height"
var defaultfields = "ctime,has_dirs,mtime,readable,size,type,writable"
var memberfields = "members,members.ctime,members.has_dirs,members.id,members.image.exif,members.image.height,members.image.width,members.mime_type,members.mtime,members.name,members.nmembers,members.parent_id,members.path,members.readable,members.rshare,members.size,members.type,members.writable"

type DirResponse struct {
	Meta
	Members []Meta `json:"members"`
}

func (d *DirResponse) Respond(w http.ResponseWriter, contenttype string) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(d)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

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
type Meta struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"`
	CTime    int64  `json:"ctime"`
	MTime    int64  `json:"mtime"`
	HasDirs  bool   `json:"has_dirs"`
	Readable bool   `json:"readable"`
	Writable bool   `json:"writable"`
	// default: ctime,has_dirs,mtime,readable,size,type,writable

	MIMEType string `json:"mime_type"`
	Size     uint64 `json:"size"`
	Nmembers int    `json:"nmembers"`

	ParentID string `json:"parent_id"`
	// Chash    string      `json:"chash"`
	// Mhash    string      `json:"mhash"`
	// MOhash   string      `json:"mohash"`
	// Nhash    string      `json:"nhash"`

	// Image   *drive.Image `json:"image"`
	Image *Image `json:"image"`
	// rshare
	// Rshare interface{} `json:"rshare"`
	// zone: zone.available, zone.quota, zone.used
	// Zone interface{} `json:"zone"`
}

func (m *Meta) Respond(w http.ResponseWriter, contenttype string) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(m)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

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
type Exif2 struct {
	Aperture         string
	BitsPerSample    string
	DateTimeOriginal string
	ExifImageHeight  string
	ExifImageWidth   string
	ExposureTime     string
	FocalLength      string
	ISO              string
	ImageHeight      string
	ImageWidth       string
	Make             string
	Model            string
	Orientation      string
	ResolutionUnit   string
	XResolution      string
	YResolution      string
}

type Responder interface {
	Respond(http.ResponseWriter, string)
}
