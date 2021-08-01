package hidrive

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
)

func checkHTTPError(w http.ResponseWriter, err error) bool {
	fmt.Printf("checkHTTPError: %T\n", err)
	if hiErr, ok := err.(*hidriveError); ok {
		b, _ := json.Marshal(hiErr)
		http.Error(w, string(b), hiErr.Code)
		return true
	}

	if err != nil {
		status := http.StatusInternalServerError
		cause := errors.Cause(err)
		if os.IsNotExist(cause) {
			status = http.StatusNotFound
		} else if os.IsExist(cause) {
			status = http.StatusInternalServerError
		} else if os.IsPermission(cause) {
			status = http.StatusForbidden
		} else if e, ok := cause.(*os.PathError); ok {
			switch e {

			case os.ErrClosed:
				status = http.StatusGone
			case os.ErrNoDeadline:
				status = http.StatusInternalServerError
			}

			//http.Error(w, fmt.Sprintf("---%v %v %v", e.Op, e.Path, e.Err.Error()), 500)
		}
		http.Error(w, cause.Error(), status)
		return true
	}
	return false
}

func DispatchRaw(wd hiDrive) http.HandlerFunc {

	var dispatchRaw http.HandlerFunc

	dispatchRaw = func(w http.ResponseWriter, r *http.Request) {

		cleanedPath := path.Clean(path.Join("/users/ihleven", strings.Replace(r.URL.Path, "|", ".", 1)))
		fmt.Println(cleanedPath)

		err := hidriveStreamFile(cleanedPath, wd.Token.GetAccessToken(), w)
		fmt.Println("ERROR:", err)
		if checkHTTPError(w, err) {
			return
		}

	}
	return dispatchRaw
}
