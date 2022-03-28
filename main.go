package main

import (
	"embed"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/ihleven/pkg/errors"
	"github.com/ihleven/pkg/httpauth"
	"github.com/ihleven/pkg/httpsrvr"
	"github.com/ihleven/pkg/log"
)

func main() {

	var usermap = map[string]string{
		"matt":     "$2a$14$zwGqBhhCzCQMBKV3zlDO5.f1FGCUSYDIiN6D9aTY2yJ5DLmJ1TsdW",
		"wolfgang": "$2a$14$KWkdJOJLa4FkKHyXZ9xFceutb8qqkQ0V2Ue1.Ce9Rn0OD69.tDHHC",
	}

	hauth := httpauth.NewAuth(usermap, []byte("my_secret_key"))

	log.Info("Hallo Welt!", log.String("asdf", "sdf"), log.String("foo", "bar"))
	// log.Info("asdf%s", 89).Err(errors.New("sdfsaf")).Package("Asdf").Msg("ASdf")
	// log.Int("asdf", 56).Info("asdf%s", 89)
	// log.Setup()

	srv := httpsrvr.NewServer(8001, true, httpsrvr.WithAuth(hauth), httpsrvr.WithEncryptecSession("SESSION", []byte("SESSION_KEY"), []byte("ENCRYPTION_KEY16")))

	srv.Register("/test", func(w http.ResponseWriter, r *http.Request) error {

		password := r.URL.Query().Get("password")
		hash, err := httpauth.HashPassword(password)
		if err != nil {
			return err
		}
		w.Write([]byte(hash))
		return nil
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
	srv.Register("/login", LoginFormHandler(hauth))
	srv.Register("/welcome", WelcomeHandler)

	// log.Errorf(nil, " === test logErrorf === %s %d", "id", 78)
	// log.Debugf(" === test log.Infof === %d %s", 17, "wach")
	// log.Info(" === test log.Info:", log.Err(errors.New("pkgerr")))
	// log.Info("test log.Info:", log.Int("number", 17), log.Str("hallo", "wach"))
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

	apirtr := httpsrvr.NewShiftPathRouter(http.NotFoundHandler(), "root")
	apirtr.Register("GET", "/bilder/(?P<id>[0-9]+$)", func(w *httpsrvr.ResponseWriter, r *http.Request, p map[string]string) error {
		// fmt.Fprintln(w, "exec", r.URL.Path, p)
		// return errors.New("asdfasdfasdfasdfasf")
		return fmt.Errorf("neuer Fejler")
	})

	return func(rw *httpsrvr.ResponseWriter, r *http.Request) error {
		// return errors.Wrap(errors.New("asdfasdfasdfasdfasf"), "asdf")
		return fmt.Errorf("neuer Fejler")
	}
}

//go:embed templates/*
var templates embed.FS

func LoginFormHandler(a *httpauth.Auth) func(w *httpsrvr.ResponseWriter, r *http.Request) {

	return func(w *httpsrvr.ResponseWriter, r *http.Request) {

		data := map[string]interface{}{}

		if r.Method == http.MethodPost {

			account := a.Authenticate(r.PostFormValue("username"), r.PostFormValue("password"))
			if account != "" {
				a.Login(w, account)
				w.Session.Values["account"] = account
				err := w.Session.Save(r, w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				http.Redirect(w, r, "/welcome", 303)
				return
			}
			data["error"] = "Invalid credentials"
		}

		t, err := template.ParseFS(templates, "templates/*.html")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.WriteHeader(200)
		t.ExecuteTemplate(w, "login.html", data)

	}
}

func WelcomeHandler(w *httpsrvr.ResponseWriter, r *http.Request) {

	if w.Session.Values["account"] != nil {
		// bar := w.Session.Values["foo"]
		account := w.Session.Values["account"]
		w.Write([]byte(fmt.Sprintf("Welcome %s !!!", account.(string))))
	} else {
		// w.WriteHeader(401)
		http.Redirect(w, r, "/login", 302)
	}

}
