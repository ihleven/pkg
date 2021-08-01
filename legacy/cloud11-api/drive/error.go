package drive

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type DriveError struct {
	Code int
	//StatusText string
	error
}

func (e *DriveError) Error() string {
	return fmt.Sprintf("http %d: %v", e.Code, e.error.Error())
}

func NewError(code int, err error) *DriveError {
	return &DriveError{code, err}
}

func HandleDriveError(erro error, w http.ResponseWriter) {

	switch err := errors.Cause(erro).(type) {
	case *DriveError:
		http.Error(w, "DriveError => "+err.Error(), err.Code)
	default:
		http.Error(w, err.Error(), 500)
	}
}
