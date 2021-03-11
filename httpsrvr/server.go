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

	loglevel := log.INFO
	if debug {
		loglevel = log.DEBUG
	}
	host := ""
	return &httpServer{
		addr:   fmt.Sprintf("%s:%d", host, port),
		routes: NewDispatcher(nil),
		// systemd:   systemd,
		debug:     debug,
		log:       log.NewStdoutLogger(loglevel),
		timestamp: time.Now().Format("20060102150405"),
	}
}

type httpServer struct {
	server    *http.Server
	routes    *dispatcher
	log       logger
	addr      string
	debug     bool
	systemd   bool
	limiter   *rate.Limiter
	counter   uint64
	timestamp string
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
		Handler:        limit(s, s.limiter), // limit(s.dispatcher, s.limiter),
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

	requestCount := atomic.AddUint64(&s.counter, 1)
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("%s-%d", s.timestamp, requestCount)
	}
	rw := NewResponseWriter(w)

	defer func(start time.Time, requestCount uint64, requestID string) {

		color.Green("request %d: %s %s => %d (%d bytes, %v)\n", requestCount, requestID, r.URL.Path, rw.statusCode, rw.Count(), time.Since(start))
	}(time.Now(), requestCount, requestID)

	ctx := context.WithValue(r.Context(), "reqid", requestID)
	ctx = context.WithValue(ctx, "counter", requestCount)
	ctx = context.WithValue(ctx, "debug", s.debug)

	r2 := r.WithContext(ctx)

	s.routes.ServeHTTP(rw, r2)

	// color.Green("request %d: %s %s => %d %v\n", requestCount, requestID, r.URL.Path, rw.statusCode, time.Since(start))
}
