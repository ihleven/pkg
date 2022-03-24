package httpsrvr

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/ihleven/errors"
)

type requestid struct{}

func (s *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	start := time.Now()

	reqnum := atomic.AddUint64(&s.counter, 1)
	reqid := r.Header.Get("X-Request-ID")
	if reqid == "" {
		reqid = fmt.Sprintf("%s-%d", s.instance, reqnum)
	}

	ctx := context.WithValue(r.Context(), requestid{}, reqid)
	// ctx = context.WithValue(ctx, "counter", reqnum)
	// ctx = context.WithValue(ctx, "debug", s.debug)
	r = r.WithContext(ctx)

	rw := NewResponseWriter(w, s.debug, s.debug)

	route, tail := s.routes.Dispatch(r.URL.Path, rw.Params)
	// if !route.preserve {
	r.URL.Path = tail
	// }

	// The key point to note is that the 'f()' in 'defer f()' is not executed when the defer statement executes
	// but the expression 'e' in 'defer f(e)' is evaluated when the defer statement executes.
	defer s.LogRequest(r, reqid, reqnum, start, rw, route.Name)

	// TODO: kann das durch route.ServeHTTP ersetzt werdem?!?
	if handler, ok := route.Handler[r.Method]; ok {
		handler.ServeHTTP(rw, r)
	} else {
		route.Handler[""].ServeHTTP(rw, r)
	}

	// var err error

	// switch h := route.handler.(type) {
	// case ResponseWriterErrorHandlerFunc:
	// 	err = h(rw, r)
	// case ErrorHandlerFunc:
	// 	err = h(rw, r)
	// default:
	// route.handler.ServeHTTP(rw, r)
	// }

	// if err != nil {
	// 	rw.HandleError(err)
	// }
}

func ConvertHandlerType(customhandler interface{}) (http.Handler, error) {
	switch h := customhandler.(type) {

	// custom HandlerFunc with custom ResponseWriter
	case func(*ResponseWriter, *http.Request):
		return ResponseWriterHandlerFunc(h), nil

	// custom HandlerFunc with custom ResponseWriter & error
	case func(*ResponseWriter, *http.Request, map[string]string) error:
		return ParamsErrorHandler(h), nil

	// custom HandlerFunc with custom ResponseWriter & error
	case func(*ResponseWriter, *http.Request) error:
		return ResponseWriterErrorHandlerFunc(h), nil

	// http.HandlerFunc with error
	case func(http.ResponseWriter, *http.Request) error:
		return ErrorHandlerFunc(h), nil

	// http.HandlerFunc:
	case func(http.ResponseWriter, *http.Request):
		return http.HandlerFunc(h), nil

	case http.Handler:
		return h, nil

	}
	return nil, errors.New("Could not convert handler type %T", customhandler)
}

// The ResponseWriterHandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, ResponseWriterHandlerFunc(f) is a
// Handler that calls f.
type ResponseWriterHandlerFunc func(*ResponseWriter, *http.Request)

// ServeHTTP calls f(w, r).
func (f ResponseWriterHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(w.(*ResponseWriter), r)
}

type ResponseWriterErrorHandlerFunc func(*ResponseWriter, *http.Request) error

func (h ResponseWriterErrorHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	rw := w.(*ResponseWriter)
	err := h(rw, r)
	if err != nil {
		// rw.err = err
		// debug := r.Context().Value("debug").(bool)
		rw.RespondError(err)
	}
}

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request) error

func (h ErrorHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	err := h(w, r)
	if err != nil {
		if rw, ok := w.(*ResponseWriter); ok {
			rw.RespondError(err)
		} else {
			http.Error(w, "w is not a *httpsrvr.ResponseWriter", 500)
		}
	}
}

type ParamsErrorHandler func(*ResponseWriter, *http.Request, map[string]string) error

func (h ParamsErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	rw := w.(*ResponseWriter)
	err := h(rw, r, rw.Params)
	if err != nil {
		// rw.err = err
		// debug := r.Context().Value("debug").(bool)
		rw.RespondError(err)
	}
}
