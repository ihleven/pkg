package httpsrvr

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ihleven/pkg/httpauth"
	"github.com/ihleven/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// only available on linux, see systemd.go
var listenAndServeSystemD func(*http.Server) error


func NewServer(port int, debug bool, options ...Option) *httpServer {

	start := time.Now()
	fd, err := os.OpenFile("tmp.log", os.O_RDWR, os.ModeAppend)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	host := ""
	srvr := &httpServer{
		addr:      fmt.Sprintf("%s:%d", host, port),
		routes:    NewShiftPathRouter(http.NotFoundHandler(), "root"),
		debug:     debug,
		startedAt: start,
		instance:  start.Format("060102-150405"),
		// logger:        nil,
		requestLogger: NewZapRequestLogger(fd),
		// carrier:       httpauth.NewCookieCarrier([]byte("my_secret_key")),
		SessionStore: nil,
	}
	for _, opt := range options {
		opt(srvr)
	}
	return srvr
}

type httpServer struct {
	server    *http.Server
	routes    *ShiftPathRoute
	addr      string
	debug     bool
	systemd   bool
	limiter   *rate.Limiter
	instance  string
	counter   uint64
	startedAt time.Time
	// logger        logger
	requestLogger *zap.Logger
	auth          *httpauth.Auth
	SessionStore  *sessions.CookieStore
}

type Option func(*httpServer)

func Port(port int) func(srvr *httpServer) {
	return func(srvr *httpServer) {
		srvr.addr = fmt.Sprintf("%s:%d", "", port)
	}
}

// func Logger(logger logger) func(srvr *httpServer) {
// 	return func(srvr *httpServer) {
// 		srvr.logger = logger
// 	}
// }

// SetLimit enables rate limiting that allows events up to rate r and permits bursts of at most b tokens.
func SetLimit(r float64, bursts int) func(srvr *httpServer) {
	return func(srvr *httpServer) {
		srvr.limiter = rate.NewLimiter(rate.Limit(r), bursts)
	}
}

// WithSystemd enables or disables systemd mode
func WithSystemd(enabled bool) func(srvr *httpServer) {
	return func(srvr *httpServer) {
		srvr.systemd = enabled
	}
}

// WithSystemd enables or disables systemd mode
func WithAuth(auth *httpauth.Auth) func(srvr *httpServer) {
	return func(srvr *httpServer) {
		srvr.auth = auth
	}
}

// WithSystemd enables or disables systemd mode
func WithSession(SESSION_KEY []byte) func(srvr *httpServer) {
	return func(srvr *httpServer) {
		srvr.SessionStore = sessions.NewCookieStore(SESSION_KEY)
	}
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
		log.Infof("+++ Starting systemd http server +++")
		err = listenAndServeSystemD(s.server)
	} else {
		log.Infow("+++ Starting http server on %v +++", "addr", s.server.Addr)
		err = s.server.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		// immediately returns after shutdown
		log.Errorf(err, "Could not listen on %s", s.addr)
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

	log.Infof(" +++ Server is shutting down... waiting up to 30 secs")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.server.SetKeepAlivesEnabled(false)
	if err := s.server.Shutdown(ctx); err != nil {
		log.Infof("Could not gracefully shutdown the server: %v\n", err)
	}
	close(waitForGracefulShutdownComplete)
}

type RouteOption func(*ShiftPathRoute)

func ParseAuth(authparser authParser) RouteOption {
	return func(route *ShiftPathRoute) {
		route.OptionParseRequestAuth = authparser
	}
}

func RequireAuth(authparser authParser) RouteOption {
	return func(route *ShiftPathRoute) {
		// route.OptionRequireRequestAuth = authparser
	}
}

// Register connects given handler to given path prefix
func (s *httpServer) Register(path string, handler interface{}, options ...RouteOption) *ShiftPathRoute {

	h, err := ConvertHandlerType(handler)
	if err != nil {
		log.Infof("Could not register route '%v': unknown handler type %T", path, handler)
		os.Exit(1)
	}

	route := s.routes.Register("", path, h)
	for _, applyOption := range options {
		applyOption(route)
	}
	return route

}
