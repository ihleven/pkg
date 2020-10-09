package httpsrvr

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/log"
)

func Reqlog(name string, debug bool, logger logger, handler ErrorHandler) http.HandlerFunc {

	if logger == nil {
		logger = log.NewStdoutLogger(log.DEBUG)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		code := 200
		path := r.URL.Path

		err := handler(w, r)
		if err != nil {
			code = HandleError(w, r, err, debug)
		}
		logger.Info(" +++ %s:  %s %s -> status %d in %s", name, r.Method, path, code, time.Since(start))
		// return nil
	}
}

type ErrorHandler func(http.ResponseWriter, *http.Request) error

func (h ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	err := h(w, r)

	if err != nil {
		HandleError(w, r, err, false)
		return
	}
}

func HandleError(w http.ResponseWriter, r *http.Request, err error, debug bool) int {

	// w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// w.WriteHeader(code)
	// fmt.Fprintf(w, "{statusCode: %d, statusText: %q, message: %q}", code, http.StatusText(code), err.Error())

	code := errors.Code(err)
	if code == int(errors.NoCode) {
		code = 500
	}
	if code == math.MaxUint16 {
		code = 500
	}

	if debug {
		http.Error(w, fmt.Sprintf("%+v", err), code)
	} else {
		http.Error(w, fmt.Sprintf("%v", errors.Cause(err)), code)
	}
	return code
}
