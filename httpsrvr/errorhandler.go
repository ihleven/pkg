package httpsrvr

import (
	"fmt"
	"math"
	"net/http"

	"github.com/ihleven/errors"
)

type ErrorHandler func(http.ResponseWriter, *http.Request) error

func (h ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	debug := r.Context().Value("debug").(bool)

	err := h(w, r)
	if err != nil {
		HandleError(w, r, err, debug)
	}
}

func HandleError(w http.ResponseWriter, r *http.Request, err error, debug bool) int {

	code := errors.Code(err)
	if code == int(errors.NoCode) {
		code = 500
	}
	if code == math.MaxUint16 {
		code = 500
	}
	w.WriteHeader(code)
	if debug {
		http.Error(w, fmt.Sprintf("%+v", err), code)
	} else {
		http.Error(w, fmt.Sprintf("%v", errors.Cause(err)), code)
	}
	return code
}
