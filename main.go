package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	args "github.com/alexflint/go-arg"
	"github.com/ihleven/pkg/hidrive"
	"github.com/ihleven/pkg/httpsrvr"
	"github.com/ihleven/pkg/kunst"
)

var (
	// https://www.reddit.com/r/golang/comments/4cpi2y/question_where_to_keep_the_version_number_of_a_go/
	// https://gist.github.com/TheHippo/7e4d9ec4b7ed4c0d7a39839e6800cc16
	version      = "undefined"
	buildtime    = "undefined"
	ClientID     = ""
	ClientSecret = ""
)

var flags struct {
	Port     int    `arg:"-p" default:"8000"  help:"Port number for non systemd mode"`
	Debug    bool   `arg:"-d" default:"false"  help:"Enable debugging output"`
	Version  bool   `arg:"-v" default:"false" help:"Print version information and exit"`
	Database string `arg:"-c" default:"wi" help:"Database name and user"`
}

func main() {

	args.MustParse(&flags)

	if flags.Version {
		fmt.Println(version)
		os.Exit(0)
	}

	repo, err := kunst.NewRepo(flags.Database)
	if err != nil {
		log.Fatal("db err:", err)
	}

	oap, err := hidrive.NewOauthProvider(hidrive.AppConfig{ClientID: ClientID, ClientSecret: ClientSecret})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	hd := hidrive.NewDrive(oap, "/wolfgang-ihle", nil)
	usermap := map[string]string{
		"matt":     "$2a$14$4zu/jv7JO377BBg0k6upZ.Ul0jqO9enCBhHlAUyoKJrcySb8JMzW2",
		"wolfgang": "$2a$14$KWkdJOJLa4FkKHyXZ9xFceutb8qqkQ0V2Ue1.Ce9Rn0OD69.tDHHC",
	}

	srv := httpsrvr.NewServer("", flags.Port, false, flags.Debug, nil)

	srv.Register("/api", kunst.ApiHandler(hd, repo, usermap))
	srv.Register("/hidrive-token-callback", http.HandlerFunc(oap.HandleAuthorizeCallback))
	srv.Register("/hidrive", hidrive.HidriveHandler(hidrive.NewDrive(oap, "", nil)))
	// srv.Register("/hidrive", hidrive.HidriveHandler(hidrive.NewDrive(oap, "", nil)))

	// srv.Register("/api", namedHandler("api"))
	srv.Register("/api/v1", namedHandler("apiV1"))
	srv.Register("/", http.FileServer(http.Dir(".")))

	srv.ListenAndServe()
}

func info(w http.ResponseWriter, r *http.Request) {
	fmt.Println("info")
	w.Write([]byte("info"))
}

func namedHandler(name string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value("reqid")
		count := r.Context().Value("counter")
		// fmt.Printf("namedHandler %s => %s\n", name, r.URL.Path)
		fmt.Fprintf(w, "GET %s => %s | %v | %v", name, r.URL.Path, id, count)

	}
}
