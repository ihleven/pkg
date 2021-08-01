package hidrive

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ihleven/cloud11-api/auth"
	"github.com/ihleven/cloud11-api/drive"
	"github.com/ihleven/cloud11-api/drive/fs"
	"github.com/pkg/errors"
)

type hiHandle struct {
	// ctime,has_dirs,mtime,readable,size,type,writable
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"`
	MIMEType string `json:"mime_type"`
	Size     uint64 `json:"size"`
	Readable bool   `json:"readable"`
	Writable bool   `json:"writable"`
	CTime    int64  `json:"ctime"`
	MTime    int64  `json:"mtime"`
	HasDirs  bool   `json:"has_dirs"`
	//Members  []Member `json:"members"`
	ID      string       `json:"id"`
	Image   *drive.Image `json:"image"`
	content []byte
}

func (h *hiHandle) Mode() os.FileMode {
	return 0644
}
func (h *hiHandle) ModTime() time.Time {
	ut := time.Unix(h.MTime, 0)
	fmt.Println("ut:", ut)
	return ut
}
func (h *hiHandle) IsDir() bool {
	return h.Type == "dir"
}

func (h *hiHandle) OpenFile(flag int, perm os.FileMode) (*os.File, error) {
	return nil, nil
}
func (h *hiHandle) ReadDir(mode os.FileMode) ([]drive.Handle, error) {
	return nil, nil
}
func (h *hiHandle) ReadImage() (*drive.Image, error) {

	i := drive.Image{
		//ColorModel: config.ColorModel,
		Width:  h.Image.Width,
		Height: h.Image.Height,
		Ratio:  float64(h.Image.Height) / float64(h.Image.Width) * 100,
		//Format:     format,
		//Src:        handle.ServeURL(),
		//Name:   handle.Name(),
		//Source: prefix,
		Exif: h.Image.Exif,
		// 	metaFile *File
		// 	Title         string
		// Caption       string // a “caption” is more like a title, while the “cutline” first describes what is happening in the picture, and then explains the significance of the event depicted.
		// Cutline       string
	}
	if err := parseMeta(h.Path, &i); err != nil {
		fmt.Println("ReadImage", h.Path, err)
		return nil, errors.Wrap(err, "Error parsing meta")
	}
	return &i, nil
}

func metaFilename(path string) string {
	base := strings.TrimSuffix(path, filepath.Ext(path))
	return fmt.Sprintf("%s.txt", base)
}

func parseMeta(path string, i *drive.Image) error {

	// 200 OK, 206 Partial content
	queryParams := url.Values{
		"path": {metaFilename(path)},
	}

	request, _ := http.NewRequest("GET", "https://api.hidrive.strato.com/2.1/file", nil)
	request.URL.RawQuery = queryParams.Encode()
	request.Header.Set("Authorization", "Bearer "+HIDrive.Token.GetAccessToken())

	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		return &hidriveError{res.StatusCode, res.Status}
	}
	if res.StatusCode >= 300 {
		if res.StatusCode == 404 {
			return nil
		}
		return NewHiDriveError(res.Body, res.StatusCode, res.Status)
	}

	re := regexp.MustCompile(`(?s)(?P<Title>.*?)=+(?P<Caption>.*?)---+(?P<Cutline>.*?)---+`)

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	match := re.FindSubmatch(content)
	paramsMap := make(map[string]string)

	for i, name := range re.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = strings.TrimSpace(string(match[i]))
		}
	}
	if title, ok := paramsMap["Title"]; ok {
		i.Title = title
	}
	if caption, ok := paramsMap["Caption"]; ok {
		i.Caption = caption
	}
	if cutline, ok := paramsMap["Cutline"]; ok {
		i.Cutline = cutline
	}
	return nil
}

func (h *hiHandle) HasReadPermission(*auth.Account) bool {
	return true
}
func (h hiHandle) Read(b []byte) (n int, err error) {

	body, err := hidriveGetFileContent(h.Path, HIDrive.Token.GetAccessToken())
	copy(b, body)
	return len(body), err
}

func (h *hiHandle) Write(b []byte) (int, error) {

	_, e := hidrivePutFile(h.Path, b)
	return len(b), e
}

func (h hiHandle) ReadSeeker() (io.ReadSeeker, error) {

	buffer := make([]byte, h.Size)
	// read file content to buffer
	//file.Read(buffer)
	fileBytes := bytes.NewReader(buffer) // converted to io.ReadSeeker type
	return fileBytes, nil
}

func (h hiHandle) GuessMIME() drive.Type {

	var t = drive.Type{
		Filetype:  "",
		Mediatype: h.Type,
		Subtype:   "",
		MIME:      h.MIMEType,
		Charset:   "",
	}
	if h.Type == "dir" {
		t.Filetype = "D"
	}
	if h.Type == "file" {
		t.Filetype = "F"
		if h.MIMEType == "application/octet-stream" {
			return *fs.GetMIMEByExtension(h.Name)
		}
		media, sub := path.Split(h.MIMEType)
		t.Mediatype = strings.TrimSuffix(media, "/")
		t.Subtype = sub
	}
	return t
}

func (h hiHandle) GetPermissions(account *auth.Account) drive.Authorization { // => handle

	perm := drive.Authorization{}

	perm.R = h.Readable
	perm.W = h.Writable
	perm.X = false
	return perm
}
