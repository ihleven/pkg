package web

import (
	"fmt"
	"net/http"
)

type EHandler func(http.ResponseWriter, *http.Request) error

func (h EHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		h.Error(err, w)
	}
}
func (h EHandler) Error(err error, w http.ResponseWriter) {

	if e, ok := err.(coder); ok {

		// errors.Printf("%v+", err)
		http.Error(w, fmt.Sprintf("%+v, %d, %s", err, e.code(), e.message()), e.code())
	}
}

type coder interface {
	error
	code() int
	message() string
}
type actionError struct {
	Code    int
	err     error
	Message string
}

func (e actionError) Error() string {
	return e.err.Error()
}
func (e actionError) code() int {
	return e.Code
}
func (e actionError) message() string {
	return e.Message
}
func (e actionError) Cause() error {
	return e.err
}
