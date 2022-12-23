package httpsrvr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/gorilla/sessions"
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
	Authkey      string
	err          error
	RuntimeStack []byte
	Session      *sessions.Session
}

// NewResponseWriter wraps given http.ResponseWriter in a ResponseWriter overwriting the WriteHeader method
func NewResponseWriter(w http.ResponseWriter, debug, pretty bool, session *sessions.Session) *ResponseWriter {

	routeparams := make(map[string]string)

	return &ResponseWriter{w, debug, pretty, routeparams, 0, 0, "", nil, nil, session}
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

// IntParamErr
func (rw *ResponseWriter) IntParamErr(key string) (int, error) {
	return strconv.Atoi(rw.Params[key])
}

// IntParam
func (rw *ResponseWriter) IntParam(key string) int {
	i, _ := strconv.Atoi(rw.Params[key])
	return i
}

//
func (rw *ResponseWriter) Respond(data interface{}, err error) {

	if err != nil {
		rw.RespondError(err)
		return
	}
	err = rw.RespondJSON(data)
	if err != nil {
		rw.RespondError(err)
	}
}

func (rw *ResponseWriter) RespondJSON(data interface{}) error {

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")

	enc := json.NewEncoder(rw)
	if rw.pretty {
		enc.SetIndent("", "    ")
	}
	err := enc.Encode(data)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal response")
	}

	return nil
}

func (rw *ResponseWriter) RespondError(err error) {

	// set error for logging
	rw.err = err

	code := errors.Code(err)
	fmt.Println("RespondError", code)
	if code == 0 {
		code = 500
	}

	if rw.debug {
		http.Error(rw, fmt.Sprintf("%#v", err), code)
	} else {
		http.Error(rw, fmt.Sprintf("%s", errors.Cause(err)), code)
	}

}
