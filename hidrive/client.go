package hidrive

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/ihleven/errors"
)

// NewClient creates a new hidrive client
func NewClient(oap *OAuth2Prov, prefix string) *HiDriveClient {

	var HiDriveClient = HiDriveClient{
		Client: http.Client{
			Timeout: 10 * time.Second,
		},
		// auth:    oap, // NewAuthProvider("", ""),
		baseURL: "https://api.hidrive.strato.com/2.1",
		// prefix:  prefix,
	}
	return &HiDriveClient
}

type HiDriveClient struct {
	http.Client
	// auth    *OAuth2Prov
	baseURL string
	// prefix  string
}

func (c *HiDriveClient) Request(method, endpoint string, params url.Values, body io.Reader, token string) (io.ReadCloser, error) {

	url := c.baseURL + endpoint + "?" + params.Encode()

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, &HiDriveError{ECode: 500, EMessage: "Failed to create new http request"}
	}
	request.Header.Set("Authorization", "Bearer "+token) // c.auth.GetAccessToken(authkey))

	response, err := c.Client.Do(request)
	if err != nil {
		if os.IsTimeout(err) {
			return nil, &HiDriveError{ECode: 500, EMessage: "client timeout exceeded"}
		}
		return nil, &HiDriveError{ECode: 500, EMessage: "HTTP client couldn't Do request"}
	}

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
		if err := NewHiDriveError(response.StatusCode, response.Body); err != nil {
			return nil, err
		}
		return nil, errors.New("HiDrive Error")
	}
}
