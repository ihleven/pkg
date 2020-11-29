package hidrive

import (
	"encoding/json"
	"net/url"

	"github.com/ihleven/errors"
)

type Meta struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"`
	MIMEType string `json:"mime_type"`
	// default: ctime,has_dirs,mtime,readable,size,type,writable
	CTime    int64 `json:"ctime"`
	MTime    int64 `json:"mtime"`
	Readable bool  `json:"readable"`
	Writable bool  `json:"writable"`

	Size     uint64 `json:"size"`
	Nmembers int    `json:"nmembers"`
	HasDirs  bool   `json:"has_dirs"`

	ParentID string `json:"parent_id"`
	// Chash    string      `json:"chash"`
	// Mhash    string      `json:"mhash"`
	// MOhash   string      `json:"mohash"`
	// Nhash    string      `json:"nhash"`

	// Image   *drive.Image `json:"image"`
	Image interface{} `json:"image"`
	// rshare
	// Rshare interface{} `json:"rshare"`
	// zone: zone.available, zone.quota, zone.used
	// Zone interface{} `json:"zone"`
}

type DirResponse struct {
	Meta
	Members []Meta `json:"members"`
}

// func (d *Meta) MarshalJSON() ([]byte, error) {
// 	type Alias Meta
// 	return json.Marshal(&struct {
// 		*Alias
// 		MTime time.Time `json:"mtime"`
// 	}{
// 		MTime: time.Unix(d.MTime, 0),
// 		Alias: (*Alias)(d),
// 	})
// }

var dirfields = "id,name,path,type,mime_type,ctime,mtime,readable,writable,size,nmembers,has_dirs,parent_id,rshare,shareable,teamfolder"
var metafields = "id,name,path,type,mime_type,ctime,mtime,readable,writable,size,nmembers,has_dirs,parent_id"
var extrafields = "rshare,shareable,teamfolder,zone"
var imagefields = "image.exif,image.width,image.height"
var defaultfields = "ctime,has_dirs,mtime,readable,size,type,writable"
var memberfields = "members,members.ctime,members.has_dirs,members.id,members.image.exif,members.image.height,members.image.width,members.mime_type,members.mtime,members.name,members.nmembers,members.parent_id,members.path,members.readable,members.rshare,members.size,members.type,members.writable"

func (c *HiDriveClient) GetMeta(path string) (*Meta, error) {

	params := url.Values{
		"path":   {path},
		"fields": {metafields},
	}
	body, hderr := c.GetReadCloser("/meta", params)
	if hderr != nil {
		return nil, hderr
	}
	defer body.Close()

	var result Meta
	err := json.NewDecoder(body).Decode(&result)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't decode response body")
	}

	return &result, nil
}

func (c *HiDriveClient) GetDir(path string, params url.Values) (*DirResponse, error) {
	if params == nil {
		params = make(map[string][]string)
	}
	memberfields := "members,members.id,members.name,members.nmembers,members.size,members.type,members.mime_type,members.mtime,members.image.height,members.image.width"
	params["path"] = []string{path}
	params["members"] = []string{"all"}
	params["fields"] = []string{metafields + "," + memberfields}

	body, err := c.GetReadCloser("/dir", params)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	var response DirResponse
	err = json.NewDecoder(body).Decode(&response)
	if err != nil {
		return nil, err
	}
	// bytes, _ := json.MarshalIndent(response, "", "    ")
	return &response, nil
}
