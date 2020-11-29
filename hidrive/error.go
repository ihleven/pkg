package hidrive

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
)

func NewHiDriveError(statusCode int, body io.Reader) *HiDriveError {

	content, err := ioutil.ReadAll(body)
	if err != nil {
		fmt.Printf("error readall NewHidriveError => %v\n", err)
	}

	var e struct {
		Code    interface{} `json:"code"`
		Message string      `json:"msg"`
	}
	err = json.Unmarshal(content, &e)
	if err != nil {
		fmt.Printf("error unmarshal NewHidriveError => %v\n", err)

	}
	var hidriveError HiDriveError
	// fmt.Printf("hidrive error: %t\n", e.Code)

	switch te := e.Code.(type) {
	case int:
		// fmt.Println("hidrive int error:", e, te)
		hidriveError = HiDriveError{HTTPStatus: statusCode, ECode: te, EMessage: e.Message}
	case float64:
		// fmt.Println("hidrive float64 error:", e, te)
		hidriveError = HiDriveError{HTTPStatus: statusCode, ECode: int(te), EMessage: e.Message}
	case string:
		i, _ := strconv.Atoi(te)
		// fmt.Println("hidrive strnerror:", e, i, e4)
		hidriveError = HiDriveError{HTTPStatus: statusCode, ECode: i, EMessage: e.Message}
	default:
		hidriveError = HiDriveError{HTTPStatus: statusCode, ECode: 0, EMessage: e.Message}
	}
	var noRegularFile = regexp.MustCompile(`not a regular file`)

	if statusCode == 403 && noRegularFile.MatchString(hidriveError.EMessage) {
		return &HiDriveError{0, 666, "noRegularFile"}
	}
	return &hidriveError
}

type HiDriveError struct {
	HTTPStatus int
	// Status     string
	ECode    int    `json:"code"`
	EMessage string `json:"msg"`
}

func (e *HiDriveError) Error() string {
	return fmt.Sprintf("%v, %s", e.ECode, e.EMessage)
}

func (e *HiDriveError) Code() int {
	return e.ECode
}
func (e *HiDriveError) HTTPStatusCode() int {
	// i, _ := strconv.Atoi(e.Code)
	return e.ECode
}

func NewAuthError(body io.Reader) (*HiDriveError, error) {

	type authError struct {
		ErrorDescription string `json:"error_description"`
		Error            string `json:"error"`
	}
	var authErr authError
	err := json.NewDecoder(body).Decode(body)
	if err != nil {

	}
	switch authErr.Error {
	case "error":
		return &HiDriveError{ECode: 401, EMessage: fmt.Sprintf("%s: %s", authErr.Error, authErr.ErrorDescription)}, nil
	case "invalid_request":
		return &HiDriveError{ECode: 400, EMessage: authErr.ErrorDescription}, nil
	}
	return nil, nil
}
