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

func NewServer(host string, port int, systemd bool, debug bool, logger logger) *httpServer {

	loglevel := log.INFO
	if logger == nil {
		if debug {
			loglevel = log.DEBUG
		}
	}
	return &httpServer{
		addr:      fmt.Sprintf("%s:%d", host, port),
		routes:    NewDispatcher(),
		systemd:   systemd,
		debug:     debug,
		log:       log.NewStdoutLogger(loglevel),
		timestamp: time.Now().Format("20060102150405"),
	}
}

type httpServer struct {
	addr      string
	debug     bool
	server    *http.Server
	systemd   bool
	limiter   *rate.Limiter
	routes    *route
	log       logger
	timestamp string
	counter   uint64
}

// NewLimiter returns a new Limiter that allows events up to rate r and permits bursts of at most b tokens.
func (s *httpServer) SetLimit(r float64, bursts int) {

	s.limiter = rate.NewLimiter(rate.Limit(r), bursts)
}

func (s *httpServer) ListenAndServe() {

	s.server = &http.Server{
		Addr:           s.addr,
		Handler:        limit(s, s.limiter), // limit(s.dispatcher, s.limiter),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    15 * time.Second, // TODO: was ist das?
		MaxHeaderBytes: 1 << 20,
	}

	// if s.limiter != nil {
	// 	s.server.Handler = limit(s.dispatcher, s.limiter)
	// }

	waitForGracefulShutdownComplete := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	// SIGTERM ist das Default-Termination-Signal von Systemd
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)

	go s.ShutdownWaiter(quit, waitForGracefulShutdownComplete)

	var err error
	if listenAndServeSystemD != nil && s.systemd {
		s.log.Info("+++ Starting systemd http server +++")
		err = listenAndServeSystemD(s.server)
	} else {
		s.log.Info("+++ Starting local http server on %v +++", s.server.Addr)
		err = s.server.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		// immediately returns after shutdown
		s.log.Fatal(err, "Could not listen on %s", s.addr)
	}

	<-waitForGracefulShutdownComplete
}

func (s *httpServer) ShutdownWaiter(quit <-chan os.Signal, waitForGracefulShutdownComplete chan<- bool) {
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

func (s *httpServer) Register(route string, handler interface{}) {

	//
	var h http.Handler
	switch handlerType := handler.(type) {
	case http.Handler:
		// s.routes.Register(route, handlerType)
		h = handlerType

	case func(w http.ResponseWriter, r *http.Request):
		// s.routes.Register(route, http.HandlerFunc(handlerType))
		h = http.HandlerFunc(handlerType)
	// case ErrorHandler:
	// s.dispatcher.Register(route, middleware(s.dispatcher.debug, s.log, handlerType))
	case func(http.ResponseWriter, *http.Request) error:
		// s.dispatcher.Register(route, middleware(s.debug, s.log, ErrorHandler(handlerType)))
		// s.routes.Register(route, ErrorHandler(handlerType))
		h = ErrorHandler(handlerType)
	// case func(*http.Request) (interface{}, error):
	// 	s.dispatcher.Register(route, ADRHandler(handlerType))

	default:
		s.log.Info("Could not register route '%v': unknown handler type %T", route, handler)
		os.Exit(1)
	}
	if route == "/" {
		s.routes.handler = h
	} else {
		s.routes.Register(route, h)
	}
}

func (s *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	start := time.Now()
	requestCount := atomic.AddUint64(&s.counter, 1)
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = fmt.Sprintf("%s-%d", s.timestamp, requestCount)
	}
	ctx := context.WithValue(r.Context(), "reqid", reqID)
	ctx2 := context.WithValue(ctx, "counter", requestCount)

	rw := NewResponseWriter(w)
	r2 := r.WithContext(ctx2)

	handler, route := s.routes.GetHandler(r.URL.Path)
	r2.URL.Path = route
	handler.ServeHTTP(rw, r2)

	color.Green("request %d: %s %s => %d %v\n", requestCount, reqID, r.URL.Path, rw.statusCode, time.Since(start))
}
