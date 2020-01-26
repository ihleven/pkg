package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/ihleven/goodies/web"
	"github.com/pkg/errors"
)

func main() {

	srv := web.NewHTTPServer()

	// dispatcher := web.NewDispatcher(http.HandlerFunc(notFoundHandler))
	srv.Register("asdf", action)
	srv.Register("foo", fooHandler)
	srv.Register("api/v1", action)
	srv.Register("api/v2", barHandler)
	srv.Register("api/v2/asdf", adr)

	srv.Run()
}

func action(w http.ResponseWriter, r *http.Request) error {

	type res struct{ Bar, Foo string }
	resa := res{Foo: "matthias", Bar: r.URL.Path}

	bytes, _ := json.MarshalIndent(resa, "", "  ")
	// if err != nil {
	// 	return err
	// }
	w.Write(bytes)

	return nil
}

func adr(r *http.Request) (interface{}, error) {

	type res struct{ Bar, Foo string }
	resa := res{Foo: "matthias", Bar: "ihle"}

	return resa, nil
}

func fooHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("foo"))
}

func barHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("foobar"))
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("notfound"))
}

func foo() error {
	return errors.WithMessage(bar(), "wrapped in foo with integer")
}

func bar() error {
	return errors.WithMessage(makeError(), "wrapped in bar")
}

func makeError() error {
	return errors.Wrap(sql.ErrNoRows, "creating error")
}
