package httpsrvr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/ihleven/pkg/errors"
)

// ResponseWriter intercepts http.ResponseWriter  capturing the response status code
type ResponseWriter struct {
	http.ResponseWriter
	debug, pretty bool
	Params        map[string]string
	count         uint64
	statusCode    int
	// status int
	// length int
	err error
}

// NewResponseWriter wraps given http.ResponseWriter in a ResponseWriter overwriting the WriteHeader method
func NewResponseWriter(w http.ResponseWriter, debug, pretty bool) *ResponseWriter {

	routeparams := make(map[string]string)

	return &ResponseWriter{w, debug, pretty, routeparams, 0, 0, nil}
}

// Write captures the response size
func (rw *ResponseWriter) Write(buf []byte) (int, error) {
	// WriteHeader(int) is not called if our response implicitly returns 200 OK, so we default to that status code.
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(buf)
	atomic.AddUint64(&rw.count, uint64(n))
	return n, err
}

// WriteHeader captures the response status
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Count function return counted bytes
func (rw *ResponseWriter) Count() uint64 {
	return atomic.LoadUint64(&rw.count)
}

//
func (rw *ResponseWriter) RespondJSON(data interface{}) {

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	var bytes []byte
	var err error
	if rw.pretty {
		bytes, err = json.MarshalIndent(data, "", "    ")
	} else {
		bytes, err = json.Marshal(data)
	}

	if err != nil {
		rw.RespondError(err)
		return
	}
	rw.Write(bytes)
}

func (rw *ResponseWriter) RespondError(err error) {

	rw.err = err

	var code int

	type coder interface{ Code() int }
	if errWithCode, ok := err.(coder); ok {
		code = errWithCode.Code()
	}

	if code == 0 {
		code = 500
	}

	rw.WriteHeader(code)

	if rw.debug {
		http.Error(rw, fmt.Sprintf("%+v", err), code)
	} else {
		http.Error(rw, fmt.Sprintf("%v", errors.Cause(err)), code)
	}

}
