package hidrive

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/ihleven/errors"
)

type Status int

const OK Status = 200
const BadRequest Status = 400   // (e.g. invalid parameter)
const Unauthorized Status = 401 // (no authentication)
const Forbidden Status = 403    // (e.g. forbidden: acl)
const NotFound Status = 404     // (e.g. file not existing)
const InternalError Status = 500

func NewHidriveError(response *http.Response) *HidriveError {

	var e struct {
		Code    interface{} `json:"code"`
		Message string      `json:"msg"`
	}

	err := json.NewDecoder(response.Body).Decode(&e)
	if err != nil {
		return &HidriveError{response.StatusCode, 0, err.Error()}
	}

	hderr := HidriveError{Status: response.StatusCode, ErrorMsg: e.Message}

	switch code := e.Code.(type) {
	case int:
		hderr.ErrorCode = code
	case float64:
		hderr.ErrorCode = int(code)
	case string:
		hderr.ErrorCode, _ = strconv.Atoi(code)
	}

	var noRegularFile = regexp.MustCompile(`not a regular file`)

	if response.StatusCode == 403 && noRegularFile.MatchString(hderr.ErrorMsg) {
		fmt.Printf("err %#v\n", hderr)
		return &HidriveError{403, 666, "noRegularFile"}
	}
	return &hderr
}

type HidriveError struct {
	Status    int    `json:"status"`
	ErrorCode int    `json:"code"`
	ErrorMsg  string `json:"msg"`
}

func (e *HidriveError) Error() string {
	return fmt.Sprintf("%v, %s", e.ErrorCode, e.ErrorMsg)
}

func (e *HidriveError) Code() int {
	return e.ErrorCode
}
func (e *HidriveError) HTTPStatusCode() int {
	// i, _ := strconv.Atoi(e.Code)
	return e.ErrorCode
}

// func NewAuthError(body io.Reader) (*HidriveError, error) {

// 	type authError struct {
// 		ErrorDescription string `json:"error_description"`
// 		Error            string `json:"error"`
// 	}
// 	var authErr authError
// 	err := json.NewDecoder(body).Decode(body)
// 	if err != nil {

// 	}
// 	switch authErr.Error {
// 	case "error":
// 		return &HiDriveError{ECode: 401, EMessage: fmt.Sprintf("%s: %s", authErr.Error, authErr.ErrorDescription)}, nil
// 	case "invalid_request":
// 		return &HiDriveError{ECode: 400, EMessage: authErr.ErrorDescription}, nil
// 	}
// 	return nil, nil
// }

func NewOAuth2Error(response *http.Response) error {
	e := OAuth2Error{Status: response.StatusCode}
	err := json.NewDecoder(response.Body).Decode(&e)
	if err != nil {
		return errors.NewWithCode(errors.ErrorCode(response.StatusCode), err.Error())
	}
	return &e
}

type OAuth2Error struct {
	Status           int    `json:"-"`
	ErrorMessage     string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e *OAuth2Error) Error() string {
	return fmt.Sprintf("Error (Status: %d) %s => %s", e.Status, e.ErrorMessage, e.ErrorDescription)
}

func (e *OAuth2Error) Code() int {
	return e.Status
}
