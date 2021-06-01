package httpsrvr

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/ihleven/pkg/log"
	"golang.org/x/time/rate"
)

// only available on linux, see systemd.go
var listenAndServeSystemD func(*http.Server) error

func NewServer(port int, debug bool) *httpServer {

	start := time.Now()

	loglevel := log.INFO
	if debug {
		loglevel = log.DEBUG
	}
	host := ""
	return &httpServer{
		addr:   fmt.Sprintf("%s:%d", host, port),
		routes: NewDispatcher(nil, "root"),
		// systemd:   systemd,
		debug:     debug,
		log:       log.NewStdoutLogger(loglevel),
		logger:    log.AccessLogger{"CombineLoggerType"}, // log.NewStdoutLogger(loglevel),
		startedAt: start,
		instance:  start.Format("20060102T150405"),
	}
}

type httpServer struct {
	server    *http.Server
	routes    *dispatcher
	log       logger
	logger    accesslogger
	addr      string
	debug     bool
	systemd   bool
	limiter   *rate.Limiter
	instance  string
	counter   uint64
	startedAt time.Time
}

// NewLimiter returns a new Limiter that allows events up to rate r and permits bursts of at most b tokens.
func (s *httpServer) SetLimit(r float64, bursts int) *httpServer {

	s.limiter = rate.NewLimiter(rate.Limit(r), bursts)
	return s
}

// WithSystemd enables or disables systemd mode
func (s *httpServer) WithSystemd(enabled bool) *httpServer {

	s.systemd = enabled
	return s
}

// WithSystemd enables or disables systemd mode
func (s *httpServer) SetLogger(logger logger) *httpServer {

	s.log = logger
	return s
}

func (s *httpServer) ListenAndServe(host string, port int) {
	s.addr = fmt.Sprintf("%s:%d", host, port)
	s.Run()
}

func (s *httpServer) Run() {

	s.server = &http.Server{
		Addr:           s.addr,
		Handler:        limit(s, s.limiter),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    15 * time.Second, // TODO: was ist das?
		MaxHeaderBytes: 1 << 20,
	}

	// for shutdown waiter to signal completed shutdown
	waitForGracefulShutdownComplete := make(chan bool, 1)

	go s.shutdownWaiter(waitForGracefulShutdownComplete)

	var err error
	if listenAndServeSystemD != nil && s.systemd {
		s.log.Info("+++ Starting systemd http server +++")
		err = listenAndServeSystemD(s.server)
	} else {
		s.log.Info("+++ Starting http server on %v +++", s.server.Addr)
		err = s.server.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		// immediately returns after shutdown
		s.log.Fatal(err, "Could not listen on %s", s.addr)
	}

	<-waitForGracefulShutdownComplete
}

// ShutdownWaiter waits for shutdown signal on channel {quit}.
// It then shuts down the server waiting 30 seconds for graceful shutdown.
// After that the waitForGracefulShutdownComplete channel is closed signalling the waiting ListenAndServe routine to end.
func (s *httpServer) shutdownWaiter(waitForGracefulShutdownComplete chan<- bool) {

	// for shutdown waiter to come into action
	quit := make(chan os.Signal, 1)
	// SIGTERM ist das Default-Termination-Signal von Systemd
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)

	// Warten auf SIGTERM
	<-quit

	s.log.Info(" +++ Server is shutting down... waiting up to 30 secs")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.server.SetKeepAlivesEnabled(false)
	if err := s.server.Shutdown(ctx); err != nil {
		s.log.Info("Could not gracefully shutdown the server: %v\n", err)
	}
	close(waitForGracefulShutdownComplete)
}

// Register connects given handler to given path prefix
func (s *httpServer) Register(path string, handler interface{}) *dispatcher {

	switch h := handler.(type) {
	case http.Handler:
		return s.routes.Register(path, h)

	case func(w http.ResponseWriter, r *http.Request):
		return s.routes.Register(path, http.HandlerFunc(h))

	case ErrorHandler:
		return s.routes.Register(path, h)

	case func(http.ResponseWriter, *http.Request) error:
		return s.routes.Register(path, ErrorHandler(h))

	default:
		s.log.Info("Could not register route '%v': unknown handler type %T", path, handler)
		os.Exit(1)
	}
	return nil
}

func (s *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	start := time.Now()
	reqnum := atomic.AddUint64(&s.counter, 1)
	reqid := r.Header.Get("X-Request-ID")
	if reqid == "" {
		reqid = fmt.Sprintf("%s-%d", s.instance, reqnum)
	}

	rw := NewResponseWriter(w)

	ctx := context.WithValue(r.Context(), "reqid", reqid)
	ctx = context.WithValue(ctx, "counter", reqnum)
	ctx = context.WithValue(ctx, "debug", s.debug)

	r = r.WithContext(ctx)

	dispatcher, tail := s.Dispatch(r.URL.Path)
	if !dispatcher.preserve {
		r.URL.Path = tail
	}

	defer func(start time.Time, reqnum uint64, reqid string, name string) {
		// err := recover()
		// if err != nil {
		// 	color.Red(" error request %d: %s %s => %d (%d bytes, %v)\n", reqnum, reqid, r.URL.Path, rw.statusCode, rw.Count(), time.Since(start))
		// 	color.Red(" recover from panic  =>  %+v\n", err)
		// }

		s.logger.Access(reqnum, reqid, start, r.RemoteAddr, "user", r.Method, r.URL.Path, r.Proto, rw.statusCode, int(rw.Count()), time.Since(start), r.Referer(), name)
		color.Green("request %d: %s %s => %d (%d bytes, %v)\n", reqnum, reqid, r.URL.Path, rw.statusCode, rw.Count(), time.Since(start))
	}(start, reqnum, reqid, dispatcher.name)

	dispatcher.handler.ServeHTTP(rw, r)
}

func (s *httpServer) Dispatch(route string) (*dispatcher, string) {

	head, tail := shiftPath(route)

	if disp, ok := s.routes.children[head]; ok {
		return disp.GetDispatcher(tail)
	}
	return s.routes, route
}
