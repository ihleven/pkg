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
		if err != nil {
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

func (c *HiDriveClient) DeleteFile(path, pid string, parentMTime int, token string) error {

	var query url.Values
	if path != "" {
		query["path"] = []string{path}
	}
	if pid != "" {
		query["pid"] = []string{pid}
	}
	if parentMTime != 0 {
		query["parent_mtime"] = []string{strconv.Itoa(parentMTime)}
	}
	_, err := c.Request("DELETE", "/file", query, nil, token)
	if err != nil {
		return errors.Wrap(err, "Error in DeleteFile request")
	}

	return nil
}

func (c *HiDriveClient) GetThumb(path, pid string, token string) (io.ReadCloser, error) {

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

func (c *HiDriveClient) GetMeta(path, pid, fields string, token string) (*FileDirSymlinkMeta, error) {

	params := url.Values{
		"path":   {path},
		"fields": {metafields + "," + imagefields},
	}

	body, err := c.Request("GET", "/meta", params, nil, token)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t execute hidrive meta request")
	}
	defer body.Close()

	var meta FileDirSymlinkMeta
	err = json.NewDecoder(body).Decode(&meta)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't decode response body")
	}

	return &meta, nil
}

func (c *HiDriveClient) GetDir(path, pid, members string, offset, limit int, fields, sort string, token string) (*Dir, error) {

	params := map[string][]string{"path": {path}, "members": {"all"}}
	if members != "" {
		params["members"] = []string{members}
	}
	if offset != 0 || limit != 0 {
		params["limit"] = []string{strconv.Itoa(offset) + "," + strconv.Itoa(limit)}
	}

	// memberfields := "members,members.id,members.name,members.nmembers,members.size,members.type,members.mime_type,members.mtime,members.image.height,members.image.width,members.image.exif"
	memberfields := "members.name,members.size,members.type,members.mime_type"
	params["fields"] = []string{memberfields}

	body, err := c.Request("GET", "/dir", params, nil, token)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t execute hidrive api request")
	}
	defer body.Close()

	var response Dir
	err = json.NewDecoder(body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *HiDriveClient) PostDir(path, pid, on_exist string, mtime, parent_mtime int, token string) (*Base, error) {

	params := url.Values{
		"path":     {path},
		"on_exist": {"autoname"},
	}

	body, err := c.Request("POST", "/dir", params, nil, token)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t execute hidrive create dir request")
	}
	defer body.Close()

	var response Base
	err = json.NewDecoder(body).Decode(&response)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding hidrive create dir response")
	}
	fmt.Printf("PostDir: %s -> %v", path, response)
	return &response, nil
}
