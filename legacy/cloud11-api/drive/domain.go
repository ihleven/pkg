package drive

import (
	"image/color"
	"io"
	"os"
	"time"

	"github.com/ihleven/cloud11-api/auth"
)

// WebDrive is the domain and can modify state, interacting with storage and/or manipulating data as needed.
// It contains the business logic.

type PathType int

const (
	DrivePath PathType = iota
	URLPath
	AbsPath
)

type Driver interface {
	Open(string) (Handle, error)
	OpenFile(string, *auth.Account) (*File, error)
	Create(string, PathType) (Handle, error)
	Mkdir(string) (Handle, error)
	ListFiles(*File, *auth.Account) ([]File, error)
	// CreateFile(folder *File, name string) (Handle, error)
	GetHandle(string, PathType) (Handle, error)
}

type Handle interface {
	//Name() string       // base name of the file
	//Size() int64        // length in bytes for regular files; system-dependent for others
	//Mode() os.FileMode  // file mode bits
	ModTime() time.Time // modification time
	IsDir() bool        // abbreviation for Mode().IsDir()
	//Sys() interface{}   // underlying data source (can return nil)
	OpenFile(flag int, perm os.FileMode) (*os.File, error)
	ReadDir(mode os.FileMode) ([]Handle, error)
	ReadImage() (*Image, error)
	io.Reader
	io.Writer
	//io.Seeker
	//io.Closer

	HasReadPermission(*auth.Account) bool
}

// //FileResponder builds the entire HTTP response from the domain's output which is given to it by the action.
// type FileResponder struct {
// 	handle Handle
// }

// File bundles all publically available information about Files (and Folders).
type File struct {
	Handle        `json:"-"`
	URL           string        `json:"url"`
	Name          string        `json:"name"`
	Size          int64         `json:"size"`
	Mode          os.FileMode   `json:"mode"`
	Type          Type          `json:"type"`
	Permissions   string        `json:"permissions"`
	Owner         *User         `json:"owner"`
	Group         *Group        `json:"group"`
	Authorization Authorization `json:"auth"`
	//Created     *time.Time   `json:"created"`
	Modified time.Time `json:"modified"`
	//Accessed    *time.Time   `json:"accessed"`
}

type Type struct {
	Filetype  string `json:"filetype"`
	Mediatype string `json:"mediatype"`
	Subtype   string `json:"subtype"`
	MIME      string `json:"mime"`
	Charset   string `json:"charset"`
}

type User struct {
	Uid      string `json:"uid"`
	Gid      string `json:"-"`
	Username string `json:"name"`
	Name     string `json:"-"`
	HomeDir  string `json:"-"`
}

type Group struct {
	Gid  string `json:"gid"`  // group ID
	Name string `json:"name"` // group name
}

type Authorization struct {
	IsOwner bool `json:"isOwner"`
	InGroup bool `json:"inGroup"`
	R       bool `json:"read"`
	W       bool `json:"write"`
	X       bool `json:"exec"`
}

type Folder struct {
	*File
	Account     *auth.Account `json:"account"`
	Drive       Driver        `json:"drive"`
	Breadcrumbs []Breadcrumb  `json:"breadcrumbs"`
	Entries     []File        `json:"entries"`
}

type Breadcrumb struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type DriveAction struct {
	//Template string
	Account     *auth.Account `json:"account"`
	Drive       Driver        `json:"drive"`
	Breadcrumbs []Breadcrumb  `json:"breadcrumbs"`
	*File
	Content string `json:"content,omitempty"`
	Entries []File `json:"entries,omitempty"`
	Image   *Image `json:"image,omitempty"`
	path    string
}

type Image struct {
	ColorModel    color.Model
	Width, Height int
	Ratio         float64
	Format        string
	Title         string
	Caption       string // a “caption” is more like a title, while the “cutline” first describes what is happening in the picture, and then explains the significance of the event depicted.
	Cutline       string // the “cutline” is text below a picture, explaining what the reader is looking at

	// https://web.ku.edu/~edit/captions.html
	// https://jerz.setonhill.edu/blog/2014/10/09/writing-a-cutline-three-examples/

	// Caption als allgemeingültige "standalone" Bildunterschrift und Cutline als Verbindung zum Album (ausgewählte Bilder in Reihe?)
	Exif         *Exif
	MetaFilePath string
}

// type ExifAlt struct {
// 	Orientation *int
// 	Taken       *time.Time
// 	Lat,
// 	Lng *float64
// 	Model string
// }

type Exif struct {
	//DateTimeOriginal, Make, Model,
	// ImageWidth, ImageHeight, ExifImageWidth, ExifImageHeight,
	// Aperture, ExposureTime, ISO, FocalLength, Orientation,
	// XResolution, YResolution, ResolutionUnit, BitsPerSample,
	// GPSLatitude, GPSLongitude, GPSAltitude
	DateTimeOriginal string `json:"DateTimeOriginal"`
	Make             string `json:"Make"`
	Model            string `json:"Model"`
	ImageWidth       int    `json:"ImageWidth"`
	ImageHeight      int    `json:"ImageHeight"`
	ExifImageWidth   int    `json:"ExifImageWidth"`
	ExifImageHeight  int    `json:"ExifImageHeight"`

	XResolution    float64 `json:"XResolution"`
	YResolution    float64 `json:"YResolution"`
	ResolutionUnit int     `json:"ResolutionUnit"`
	BitsPerSample  int     `json:"BitsPerSample"`

	Aperture     float64 `json:"Aperture"`
	ExposureTime float64 `json:"ExposureTime"`
	ISO          int     `json:"ISO"`
	FocalLength  float64 `json:"FocalLength"`
	Orientation  float64 `json:"Orientation"`

	GPSLatitude  float64 `json:"GPSLatitude"`
	GPSLongitude float64 `json:"GPSLongitude"`
	GPSAltitude  float64 `json:"GPSAltitude"`
}
