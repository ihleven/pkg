package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ihleven/pkg/errors"
	"github.com/ihleven/pkg/httpsrvr"
)

func main() {

	srv := httpsrvr.NewServer(8001, true)

	// srv.Register("/", http.FileServer(http.FS(fsys))).Name("nuxt")
	// srv.Register("/api", authMiddleware(apirouter.Router()))

	// srv.Register("/api/file", api.AuthMiddlewareFunc(drive.Serve))
	// srv.Register("/api/thumbs", api.AuthMiddlewareFunc(drive.ThumbHandler))
	// srv.Register("/api/signin", auth.SigninHandler(usermap))
	// srv.Register("/api/signout", auth.SignoutHandler)
	srv.Register("/test", func(w http.ResponseWriter, r *http.Request) error {
		// return nil
		time.Sleep(5 * time.Second)
		return nil //errors.Wrap(errors.NewWithCode(400, "Fehler mit Code 400"), "Gewrappt")
	})
	srv.Register("/test/asdf", handler)
	srv.Register("/handler/func", handlerfunc)
	srv.Register("/handler/func2", handlerfunc2)
	srv.Register("/handler/func3", handlerfunc3)
	srv.Register("/country/(?P<blupp>[a-z]+$)", dynhandler)
	srv.Register("/accoms/(?P<id>[0-9]+$)", accomhandler)
	srv.Register("/country/(?P<country>[0-9]+$)/(?P<region>^[0-9]+$)", dynhandler)
	// srv.Register("/nested/:code/region/:region/aasdf", dynhandler)
	srv.Register("/kunst/", KunstAPI(true))

	srv.Run()
}

func handler(w *httpsrvr.ResponseWriter, r *http.Request) error {
	if e := r.URL.Query().Get("error"); e != "" {
		err := errors.New("asdfasdfasdfad")
		err = errors.Wrap(err, "eins")
		err = asdf(err)
		err = errors.Wrap(err, "")
		err = errors.Wrap(err, "vier")
		err = errors.Wrap(err, "fünf")
		return errors.Wrap(err, "das letzte")

	}
	data := struct {
		A, B int
		C    string
	}{4, 5, r.URL.Path}
	w.RespondJSON(&data)
	time.Sleep(1 * time.Second)
	return nil //errors.New("asdfasdfasdfad")
}

func asdf(err error) error {
	return errors.Wrap(err, "zwei")
}

func handlerfunc(w http.ResponseWriter, r *http.Request) {
	fmt.Println("path:", r.URL.Path)
	w.Write([]byte("Hallo Welt"))
	// w.(*httpsrvr.ResponseWriter).RespondError(errors.New("asdasdfasdf"))
}

func handlerfunc2(w *httpsrvr.ResponseWriter, r *http.Request) {

	w.RespondJSON(r.URL.Path)
}

func handlerfunc3(w *httpsrvr.ResponseWriter, r *http.Request) error {

	w.RespondJSON(r.URL.Path)

	return nil
}

func dynhandler(w *httpsrvr.ResponseWriter, r *http.Request, p map[string]string) error {

	w.RespondJSON("=== dynhandler ---" + r.URL.Path)
	w.RespondJSON(p)
	return nil
}

func accomhandler(w *httpsrvr.ResponseWriter, r *http.Request, p map[string]string) error {

	w.RespondJSON("accom: ")
	w.RespondJSON(p)
	return nil
}

func KunstAPI(debug bool) httpsrvr.ResponseWriterErrorHandlerFunc {

	apirtr := httpsrvr.NewShiftPathRouter(nil, "root")
	apirtr.Register("GET", "/bilder/(?P<id>[0-9]+$)", func(w *httpsrvr.ResponseWriter, r *http.Request, p map[string]string) error {
		// fmt.Fprintln(w, "exec", r.URL.Path, p)
		return errors.New("asdfasdfasdfasdfasf")
	})

	return func(rw *httpsrvr.ResponseWriter, r *http.Request) error {
		return errors.Wrap(errors.New("asdfasdfasdfasdfasf"), "asdf")
	}
}
