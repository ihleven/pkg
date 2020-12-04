package hidrive

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/ihleven/errors"
)

// NewClient creates a new hidrive client
func NewClient(oap *OAuth2Prov, prefix string) *HiDriveClient {

	var HiDriveClient = HiDriveClient{
		Client: http.Client{
			// Timeout: 10000,
		},
		auth:    oap, // NewAuthProvider("", ""),
		baseURL: "https://api.hidrive.strato.com/2.1",
		prefix:  prefix,
	}
	return &HiDriveClient
}

type HiDriveClient struct {
	http.Client
	auth    *OAuth2Prov
	baseURL string
	prefix  string
}

func (c *HiDriveClient) UploadFile(folder string, body io.Reader, name string, modtime string) (*Meta, error) {

	params := url.Values{"dir": {"/users/matt.ihle/wolfgang-ihle/" + folder}, "name": {name}, "on_exist": {"autoname"}, "mtime": {modtime}}
	url := c.baseURL + "/file?" + params.Encode()
	fmt.Printf("params: %v, url: %v\n", params, url)

	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, &HiDriveError{ECode: 500, EMessage: "Failed to create new http request"}
	}
	request.Header.Set("Authorization", "Bearer "+c.auth.GetAccessToken())
	response, err := c.Client.Do(request)
	if err != nil {
		if os.IsTimeout(err) {
			return nil, &HiDriveError{ECode: 500, EMessage: "client timeout exceeded"}
		}
		return nil, &HiDriveError{ECode: 500, EMessage: "HTTP client couldn't Do request"}
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case 200, 201:
		var meta Meta
		err = json.NewDecoder(response.Body).Decode(&meta)
		if err != nil {
			return nil, err
		}
		return &meta, nil
	case 204:
		return nil, nil

	default:
		if err := NewHiDriveError(response.StatusCode, response.Body); err != nil {
			return nil, err
		}
		return nil, errors.New("HiDrive Error")
	}
}
func (c *HiDriveClient) UploadFileNeu(folder string, body io.Reader, name string, modtime string) (*Meta, error) {

	respBody, err := c.PostRequest("/file", body, url.Values{
		"dir":      {folder},
		"name":     {name},
		"on_exist": {"autoname"},
		"mtime":    {modtime},
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error in post request")
	}
	defer respBody.Close()

	var meta Meta
	err = json.NewDecoder(respBody).Decode(&meta)
	if err != nil {
		fmt.Println("upload err:", err)
		return nil, errors.Wrap(err, "Error decoding post result")
	}
	fmt.Println("upload meta:", meta)
	return &meta, nil
}

func (c *HiDriveClient) PostRequest(endpoint string, body io.Reader, params url.Values) (io.ReadCloser, error) {

	if dir := params.Get("dir"); dir != "" {
		params.Set("dir", path.Join("/users/matt.ihle/wolfgang-ihle/", dir))
	}

	url := c.baseURL + endpoint + "?" + params.Encode()
	fmt.Println("upload url:", url)
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, &HiDriveError{ECode: 500, EMessage: "Failed to create new http request"}
	}
	request.Header.Set("Authorization", "Bearer "+c.auth.GetAccessToken())

	response, err := c.Client.Do(request)
	if err != nil {
		if os.IsTimeout(err) {
			return nil, &HiDriveError{ECode: 500, EMessage: "client timeout exceeded"}
		}
		return nil, &HiDriveError{ECode: 500, EMessage: "HTTP client couldn't Do request"}
	}

	// if response.StatusCode >= 300 {
	// 	return nil, NewHiDriveError(response.StatusCode, response.Body)
	// } else {
	// 	return response.Body, nil
	// }

	switch response.StatusCode {
	case 200, 201:
		return response.Body, nil
	case 204:
		return nil, nil

	default:
		if err := NewHiDriveError(response.StatusCode, response.Body); err != nil {
			return nil, err
		}
		return nil, errors.New("HiDrive Error")
	}
}

func (c *HiDriveClient) GetReadCloser(endpoint string, params url.Values) (io.ReadCloser, error) {

	url := c.baseURL + endpoint + "?" + params.Encode()

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &HiDriveError{ECode: 500, EMessage: "Failed to create new http request"}
	}
	request.Header.Set("Authorization", "Bearer "+c.auth.GetAccessToken())

	response, err := c.Client.Do(request)
	if err != nil {
		if os.IsTimeout(err) {
			return nil, &HiDriveError{ECode: 500, EMessage: "client timeout exceeded"}
		}
		return nil, &HiDriveError{ECode: 500, EMessage: "HTTP client couldn't Do request"}
	}

	// if response.StatusCode >= 300 {
	// 	return nil, NewHiDriveError(response.StatusCode, response.Body)
	// } else {
	// 	return response.Body, nil
	// }

	switch response.StatusCode {
	case 200, 201:
		return response.Body, nil
	case 204:
		return nil, nil

	default:
		if err := NewHiDriveError(response.StatusCode, response.Body); err != nil {
			return nil, err
		}
		return nil, errors.New("HiDrive Error")
	}
}

// func (c *HiDriveClient) BaseExec(method, route string, body io.Reader, result interface{}) error {

// 	request, err := http.NewRequest(method, route, body)
// 	if err != nil {
// 		return errors.Wrap(err, "Failed to create new http request (%s %s)", method, route)
// 	}
// 	if c.auth != nil {
// 		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.auth.GetAccessToken()))
// 	}
// 	response, err := c.Client.Do(request)
// 	if err != nil {
// 		if os.IsTimeout(err) {
// 			return errors.Wrap(err, "V5 client timeout of exceeded")
// 		}
// 		return errors.Wrap(err, "HTTP client couldn't Do request")
// 	}
// 	defer response.Body.Close() // Use defer only if http.Get is succesful.

// 	// a, _ := ioutil.ReadAll(resp.Body)
// 	// fmt.Println("400:", string(a))
// 	switch response.StatusCode {
// 	case 200, 201:
// 		err = json.NewDecoder(response.Body).Decode(result)
// 		if err != nil {
// 			body, readError := ioutil.ReadAll(response.Body)
// 			if readError != nil {
// 				return errors.Wrap(readError, "Couldn't read V5 response body")
// 			}
// 			return errors.Wrap(err, "Couldn't decode V5 response body: >>>%s<<<", body)
// 		}
// 		fallthrough
// 	case 204:
// 		return nil

// 	default:
// 		if err := NewHiDriveError(response.StatusCode, response.Body); err != nil {
// 			return err
// 		}
// 		return errors.New("HiDrive Error")
// 	}
// }
