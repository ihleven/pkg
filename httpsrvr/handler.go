package httpsrvr

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/sessions"
	"github.com/ihleven/pkg/errors"
)

type requestid struct{}

func (s *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	start := time.Now()

	reqnum := atomic.AddUint64(&s.counter, 1)
	reqid := r.Header.Get("X-Request-ID")
	if reqid == "" {
		reqid = fmt.Sprintf("%s-%d", s.instance, reqnum)
	}

	color.Cyan("request %d: %s %s\n", reqnum, reqid, r.URL.Path)

	ctx := context.WithValue(r.Context(), requestid{}, reqid)

	r = r.WithContext(ctx)

	var session *sessions.Session
	if s.SessionStore != nil {
		// Get a session. We're ignoring the error resulted from decoding an
		// existing session: Get() always returns a session, even if empty.
		session, _ = s.SessionStore.Get(r, "session-name")
	}

	rw := NewResponseWriter(w, s.debug, s.debug, session)

	route, tail := s.routes.Dispatch(r.URL.Path, rw.Params)
	// if !route.preserve {
	r.URL.Path = tail
	// }
	if route.OptionParseRequestAuth != nil {
		rw.Authkey = route.OptionParseRequestAuth.ParseRequestAuth(r)
	}

	// The key point to note is that the 'f()' in 'defer f()' is not executed when the defer statement executes
	// but the expression 'e' in 'defer f(e)' is evaluated when the defer statement executes.
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			rw.RuntimeStack = buf[:n]
			if e, ok := err.(error); ok {

				wrappederr := errors.Wrap(e, "our server got panic")
				rw.RespondError(wrappederr)
			} else {

				// fmt.Printf("recovering from err %v\n %s", err, buf)
				msg := fmt.Sprintf("our server got panic -> %s", err)
				w.Write([]byte(msg))
			}
		}
		s.LogRequest(r, reqid, reqnum, start, rw, route.Name)
	}()

	// defer s.LogRequest(r, reqid, reqnum, start, rw, route.Name)

	// TODO: kann das durch route.DispatchAndServe ersetzt werden?!?
	if handler, ok := route.Handler[r.Method]; ok {
		handler.ServeHTTP(rw, r)
	} else if handler, ok := route.Handler[""]; ok && handler != nil {
		handler.ServeHTTP(rw, r)
	} else {
		http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
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
