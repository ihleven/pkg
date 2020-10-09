package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ADRHandler func(*http.Request) (interface{}, error)

func (h ADRHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	response, err := h(r)
	if err != nil {
		h.Error(err, w)
		return
	}

	bytes, _ := json.MarshalIndent(response, "", "  ")
	w.Write(bytes)
}

func (h ADRHandler) Error(err error, w http.ResponseWriter) {

	if e, ok := err.(coder); ok {

		// errors.Printf("%v+", err)
		http.Error(w, fmt.Sprintf("%+v, %d, %s", err, e.code(), e.message()), e.code())
	}
}
