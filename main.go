package main

import (
	"fmt"
	"net/http"
	"os"

	args "github.com/alexflint/go-arg"
	"github.com/ihleven/pkg/httpsrvr"
	"github.com/ihleven/pkg/kunst"
)

var (
	// https://www.reddit.com/r/golang/comments/4cpi2y/question_where_to_keep_the_version_number_of_a_go/
	// https://gist.github.com/TheHippo/7e4d9ec4b7ed4c0d7a39839e6800cc16
	version   = "undefined"
	buildtime = "undefined"
)

var flags struct {
	Port     int    `arg:"-p" default:"8000"  help:"Port number for non systemd mode"`
	Debug    bool   `arg:"-d" default:"false"  help:"Enable debugging output"`
	Medien   string `arg:"-m" default:"false"  help:"Medien-Verzeichnis"`
	Version  bool   `arg:"-v" default:"false" help:"Print version information and exit"`
	Database string `arg:"-c" default:"wi" help:"Database name and user"`
}

func main() {

	args.MustParse(&flags)

	if flags.Version {
		fmt.Println(version)
		os.Exit(0)
	}

	if flags.Debug {
		fmt.Println("debug", flags)
	}

	srv := httpsrvr.NewServer("", flags.Port, false, false, nil)
	// handler := kunst.NewHandler()
	srv.Register("/api", kunst.KunstHandler(flags.Database, flags.Medien))
	// srv.Register("/api/upload", kunst.UploadFile(handler))
	// srv.Register("/bilder", kunst.Bilder(handler)) // template
	// srv.Register("/api/bild", kunst.BildDetail(handler))
	srv.Register("/api/media", http.StripPrefix("/", http.FileServer(http.Dir(flags.Medien))))

	srv.ListenAndServe()

}
