package web

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func NewHTTPServer() *httpServer {

	dispatcher := NewDispatcher(nil)
	server := httpServer{
		dispatcher: dispatcher,
		server: &http.Server{
			Addr:           ":8080",
			Handler:        dispatcher,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}
	return &server
}

type httpServer struct {
	dispatcher *shiftPathDispatcher
	server     *http.Server
}

func (s *httpServer) Run() {
	log.Fatal(http.ListenAndServe(":8080", s.dispatcher))
}

func (s *httpServer) Register(route string, handler interface{}) {

	switch handlerType := handler.(type) {
	case http.Handler:
		s.dispatcher.Register(route, handlerType)

	case func(w http.ResponseWriter, r *http.Request):
		s.dispatcher.Register(route, http.HandlerFunc(handlerType))

	case func(w http.ResponseWriter, r *http.Request) error:
		s.dispatcher.Register(route, EHandler(handlerType))

	case func(*http.Request) (interface{}, error):
		s.dispatcher.Register(route, ADRHandler(handlerType))

	default:
		fmt.Println(route, handler, "default", handlerType)
	}

}
