package hidrive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/pkg/errors"
)

type StratoHiDrive struct {
	token string
}

func getAccessToken(code string) (string, error) {

	tokenbytes, err := ioutil.ReadFile("./token")
	if err == nil {
		fmt.Println("read:", string(tokenbytes))
		return string(tokenbytes), nil
	}

	resp, err := hidriveOAuth2Token(code)
	if err != nil {
		fmt.Println("oauth error:", err)
	}

	err = ioutil.WriteFile("./token", []byte(resp.AccessToken), 0644)
	if err != nil {
		return resp.AccessToken, err
	}
	return resp.AccessToken, nil
}

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}

func hidriveOAuth2Token(code string) (*OAuthAccessResponse, error) {

	formData := url.Values{
		"client_id":     {"b4436f1157043c2bf8db540c9375d4ed"},
		"client_secret": {"8c5453a7264e4200ab80206658987dd8"},
		"grant_type":    {"authorization_code"},
		"code":          {code},
	}

	res, err := http.PostForm("https://my.hidrive.com/oauth2/token", formData)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var response OAuthAccessResponse
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}

// curl -X POST --data // https://my.hidrive.com/oauth2/token

type DirResponse struct {
	Path     string     `json:"path"`
	Type     string     `json:"type"`
	Size     uint64     `json:"size"`
	Readable bool       `json:"readable"`
	Writable bool       `json:"writable"`
	CTime    int64      `json:"ctime"`
	MTime    int64      `json:"mtime"`
	HasDirs  bool       `json:"has_dirs"`
	NMembers int        `json:"nmembers"`
	Members  []hiHandle `json:"members"`
}

func hidriveGetDir(path string, bearer string) (*DirResponse, error) {
	if path == "" {
		path = "/"
	}
	members := "members.id,members.mime_type,members.mtime,members.name,members.readable,members.writable,members.type,members.nmembers,members.path,members.size"
	queryParams := url.Values{
		"path":    {path},
		"members": {"all"},
		"fields":  {"ctime,has_dirs,id,mtime,readable,size,type,writable,nmembers,path," + members},
	}
	req, _ := http.NewRequest("GET", "https://api.hidrive.strato.com/2.1/dir", nil)
	req.URL.RawQuery = queryParams.Encode()
	req.Header.Set("Authorization", "Bearer "+bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, NewHiDriveError2(resp.StatusCode, resp.Status, body)
	}

	//fmt.Println("hidriveGetDirResponse:", string(body), res.Status, res.StatusCode, "adf")

	var response DirResponse
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

type hidriveMetaResponse struct {
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
}

type hidriveError struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func (e hidriveError) Error() string {
	return e.Message
}

func NewHiDriveError(txt io.ReadCloser, code int, status string) error {
	var hidriveError hidriveError
	if err := json.NewDecoder(txt).Decode(&hidriveError); err == nil {
		return &hidriveError
	}

	body, err := ioutil.ReadAll(txt)
	if err != nil {
		return err
	}
	hidriveError.Code = code
	hidriveError.Message = status
	fmt.Println("body:", body, code, status)
	return &hidriveError
}
func NewHiDriveError2(code int, status string, content []byte) error {
	var hidriveError hidriveError
	if err := json.Unmarshal(content, &hidriveError); err != nil {
		hidriveError.Code = code
		hidriveError.Message = status
	}
	return &hidriveError
}

func hidriveStreamFile(path string, bearer string, w io.Writer) error {
	// 200 OK, 206 Partial content
	queryParams := url.Values{
		"path": {path},
	}

	request, _ := http.NewRequest("GET", "https://api.hidrive.strato.com/2.1/file", nil)
	request.URL.RawQuery = queryParams.Encode()
	request.Header.Set("Authorization", "Bearer "+bearer)

	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		return &hidriveError{res.StatusCode, res.Status}
	}
	if res.StatusCode != 200 {
		fmt.Println(res.Body)
		return NewHiDriveError(res.Body, res.StatusCode, res.Status)
	}
	if _, err := io.Copy(w, res.Body); err != nil {
		fmt.Println("copy:", err)
		return err
	}
	fmt.Println("file getted")
	return nil
}
func hidriveGetFileContent(path string, bearer string) ([]byte, error) {

	// 200 OK, 206 Partial content
	queryParams := url.Values{
		"path": {path},
	}

	request, _ := http.NewRequest("GET", "https://api.hidrive.strato.com/2.1/file", nil)
	request.URL.RawQuery = queryParams.Encode()
	request.Header.Set("Authorization", "Bearer "+bearer)

	client := http.DefaultClient
	resp, err := client.Do(request)
	if err != nil {
		return nil, &hidriveError{resp.StatusCode, resp.Status}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		return body, NewHiDriveError2(resp.StatusCode, resp.Status, body)
	}
	return body, nil

}

func hidrivePutFile(p string, content []byte) (*hiHandle, error) {

	dir, name := path.Split(p)
	queryParams := url.Values{
		"dir":  {strings.TrimSuffix(dir, "/")},
		"name": {name},
	}

	request, _ := http.NewRequest("PUT", "https://api.hidrive.strato.com/2.1/file", bytes.NewReader(content))
	request.URL.RawQuery = queryParams.Encode()
	request.Header.Set("Authorization", "Bearer "+HIDrive.Token.GetAccessToken())
	request.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, &hidriveError{resp.StatusCode, resp.Status}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, NewHiDriveError2(resp.StatusCode, resp.Status, body)
	}

	var handle hiHandle
	if err = json.Unmarshal(body, &handle); err != nil {
		return nil, errors.Wrap(err, "Could not unmarshall put response")
	}
	return &handle, nil
}

func hidrivePostFile(p string, bearer string) (*hiHandle, error) {

	dirname, filename := path.Split(p)
	dirname = strings.TrimSuffix(dirname, "/")
	queryParams := url.Values{
		"dir":  {dirname},
		"name": {filename},
		//"on_exist": {"overwrite"},
	}

	request, _ := http.NewRequest("POST", "https://api.hidrive.strato.com/2.1/file", strings.NewReader(""))
	request.URL.RawQuery = queryParams.Encode()
	request.Header.Set("Authorization", "Bearer "+bearer)
	request.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		return nil, &hidriveError{res.StatusCode, res.Status}
	}

	if res.StatusCode >= 300 {
		return nil, NewHiDriveError(res.Body, res.StatusCode, res.Status)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	// fmt.Println(" * hidrivePostFileResponse:", string(body), res.Status, res.StatusCode, "adf")

	var handle hiHandle
	if err = json.NewDecoder(bytes.NewReader(body)).Decode(&handle); err != nil {
		return nil, err
	}
	fmt.Printf(" * hidrivePostFileResponse: %v\n", handle)

	return &handle, nil

}
