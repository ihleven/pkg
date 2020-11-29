package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	args "github.com/alexflint/go-arg"
	"github.com/ihleven/pkg/auth"
	"github.com/ihleven/pkg/hidrive"
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
	Medien   string `arg:"-m" default:"./temp-images"  help:"Medien-Verzeichnis"`
	Version  bool   `arg:"-v" default:"false" help:"Print version information and exit"`
	Database string `arg:"-c" default:"wi" help:"Database name and user"`
}

func main() {

	oap := hidrive.NewOauthProvider()
	// oap.RefreshToken()
	// err := oap.TokenInfo()
	// if err != nil {
	// 	fmt.Printf("%+v\n", err)
	// }

	hdclient := hidrive.NewClient(oap)
	hd := hidrive.NewDrive(oap, hidrive.PrefixPath("/users/matt.ihle/wolfgang-ihle"))
	// wdrive := hidrive.NewDrive(oap, "wolfgang-ihle")

	args.MustParse(&flags)

	if flags.Version {
		fmt.Println(version)
		os.Exit(0)
	}

	if flags.Debug {
		fmt.Println("debug", flags)
		db, err := kunst.NewRepo(flags.Database)
		if err != nil {
			log.Fatal(err)
		}
		bilder, err := db.LoadBilder()
		if err != nil {
			log.Fatal(err)
		}
		for _, bild := range bilder {
			fmt.Println(" * bild ->", bild.ID, bild.Titel)
			for _, foto := range bild.Fotos {
				fmt.Println("   * foto ->", foto)
			}
		}
		// fotos, err := db.LoadFotos()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// for _, foto := range fotos {
		// 	file, err := os.Open(path.Join(flags.Medien, foto.Path))
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// 	fmt.Println(" * foto ->", foto.ID, foto.Name, file, flags.Medien+"/thumbs/100/")
		// 	err = kunst.GenerateThumbnail100(file, foto.ID, foto.Width, foto.Height, flags.Medien+"/thumbs/100/")
		// 	if err != nil {
		// 		fmt.Println("error thumbnailing", err)
		// 	}
		// }

	}

	srv := httpsrvr.NewServer("", flags.Port, false, false, nil)

	srv.Register("/api", kunst.KunstHandler(flags.Database, flags.Medien, hdclient))
	srv.Register("/api/media", http.StripPrefix("/", http.FileServer(http.Dir(flags.Medien))))
	srv.Register("/api/hidrive", hd)
	srv.Register("/api/signin", auth.SigninHandler)

	// srv.Register("/auth", oap)
	srv.Register("/cloud11/hidrive", hd)
	srv.Register("/cloud11/signin", auth.SigninHandler)
	// srv.Register("/welcome", auth.Welcome)
	// srv.Register("/refresh", auth.Refresh)
	srv.ListenAndServe()

}
