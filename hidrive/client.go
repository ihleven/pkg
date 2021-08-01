package hidrive

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/ihleven/errors"
)

// NewClient creates a new hidrive client
func NewClient() *HiDriveClient {

	var HiDriveClient = HiDriveClient{
		Client: http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://api.hidrive.strato.com/2.1",
	}
	return &HiDriveClient
}

type HiDriveClient struct {
	http.Client
	baseURL string
}

func (c *HiDriveClient) Request(method, endpoint string, params url.Values, body io.Reader, token string) (io.ReadCloser, error) {

	url := c.baseURL + endpoint + "?" + params.Encode()

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create new hidrive client http request")
	}
	request.Header.Set("Authorization", "Bearer "+token)

	response, err := c.Client.Do(request)
	if err != nil {
		if os.IsTimeout(err) {
			return nil, errors.Wrap(err, "hidrive client request timeout exceeded") // &HiDriveError{ECode: 500, EMessage: "client timeout exceeded"}
		}
		fmt.Println(" !!! HiDrive request error: ", err)
		return nil, errors.Wrap(err, "HTTP client couldn't Do request")
	}
	// fmt.Println("client response:", response)
	switch response.StatusCode {
	// 200	OK
	// 201	Created
	// 204	No Content
	// 206	Partial Content
	// 304	Not Modified
	// 400	Bad Request (e.g. invalid parameter)
	// 401	Unauthorized (no authentication)
	// 403	Forbidden (Forbidden: Insufficient privileges)
	// 404	Not Found
	// 409	Conflict
	// 413	Request Entity Too Large
	// 415	Unsupported Media Type
	// 416	Requested Range Not Satisfiable
	// 422	Unprocessable Entity (e.g. name too long)
	// 500	Internal Error
	// 507	Insufficient Storage
	case 200, 201, 206:
		return response.Body, nil
	case 204:
		return nil, nil

	default:
		hidriveError := NewHidriveError(response)
		if hidriveError == nil {
			return nil, errors.Wrap(err, "Couldn‘t parse hidrive error")
		}
		return nil, hidriveError
	}
}

func (c *HiDriveClient) GetFile(path, pid string, token string) (io.ReadCloser, error) {

	query := url.Values{}
	if path != "" {
		query["path"] = []string{path}
	}
	if pid != "" {
		query["pid"] = []string{pid}
	}

	body, err := c.Request("GET", "/file", query, nil, token)
	if err != nil {
		return nil, errors.Wrap(err, "Error in get file request")
	}

	return body, nil
}

func (c *HiDriveClient) DeleteFileUnused(path, pid string, parentMTime int, token string) error {

	var query url.Values
	if path != "" {
		query.Set("path", path)
	}
	if pid != "" {
		query.Set("pid", pid)
	}
	if parentMTime != 0 {
		query.Set("parent_mtime", strconv.Itoa(parentMTime))
	}
	_, err := c.Request("DELETE", "/file", query, nil, token)
	if err != nil {
		return errors.Wrap(err, "Error in DeleteFile request")
	}

	return nil
}

func (c *HiDriveClient) GetThumbUnused(path, pid string, token string) (io.ReadCloser, error) {

	var query url.Values
	if path != "" {
		query["path"] = []string{path}
	}
	if pid != "" {
		query["pid"] = []string{pid}
	}

	body, err := c.Request("GET", "/file/thumbnail", query, nil, token)
	if err != nil {
		return nil, errors.Wrap(err, "Error in get file request")
	}

	return body, nil
}

func (c *HiDriveClient) GetMeta(path, pid, fields string, token string) (*Meta, error) {

	params := url.Values{
		"path":   {path},
		"fields": {metafields + "," + imagefields},
	}

	body, err := c.Request("GET", "/meta", params, nil, token)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t execute hidrive meta request")
	}
	defer body.Close()
	bytes, _ := io.ReadAll(body)
	var meta Meta
	err = json.Unmarshal(bytes, &meta)
	// err = json.NewDecoder(body).Decode(&meta)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't decode response body")
	}
	fmt.Println("meta:", string(bytes), meta.Path, meta.CTime, meta.MTime)
	return &meta, nil
}

func (c *HiDriveClient) GetDir(path, pid, members string, offset, limit int, fields, sort string, token string) (*Meta, error) {

	// var params map[string][]string
	params := map[string][]string{"path": {path}, "members": {"all"}}
	if members != "" {
		params["members"] = []string{members}
	}
	if offset != 0 || limit != 0 {
		params["limit"] = []string{strconv.Itoa(offset) + "," + strconv.Itoa(limit)}
	}

	// memberfields := "members,members.id,members.name,members.nmembers,members.size,members.type,members.mime_type,members.mtime,members.image.height,members.image.width,members.image.exif"
	memberfields := "members.id,members.name,members.path,members.size,members.nmembers,members.type,members.mime_type,members.ctime,members.mtime,members.image.height,members.image.width,members.readable,members.writable"
	params["fields"] = []string{memberfields}

	body, err := c.Request("GET", "/dir", params, nil, token)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t execute hidrive api request")
	}
	defer body.Close()

	var response Meta
	err = json.NewDecoder(body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *HiDriveClient) PostDir(path, pid, on_exist string, mtime, parent_mtime int, token string) (*Meta, error) {

	params := url.Values{
		"path":     {path},
		"on_exist": {"autoname"},
	}

	body, err := c.Request("POST", "/dir", params, nil, token)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t execute hidrive create dir request")
	}
	defer body.Close()

	var response Meta
	err = json.NewDecoder(body).Decode(&response)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding hidrive create dir response")
	}
	fmt.Printf("PostDir: %s -> %v", path, response)
	return &response, nil
}

func (c *HiDriveClient) PutFile(content io.Reader, dir, name string, mtime, parent_mtime int, token string) (*Meta, error) {

	params := url.Values{
		"dir":  {dir},
		"name": {name},
	}
	// if mtime != 0 {
	// 	params.Add("mtime", mtime)
	// }
	fmt.Println("put", params, token)
	body, err := c.Request("PUT", "/file", params, content, token)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t execute hidrive put file request")
	}
	defer body.Close()

	var response Meta
	err = json.NewDecoder(body).Decode(&response)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding hidrive put file response")
	}

	return &response, nil
}
