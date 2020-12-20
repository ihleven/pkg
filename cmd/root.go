package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ihleven/pkg/auth"
	"github.com/ihleven/pkg/hidrive"
	"github.com/ihleven/pkg/httpsrvr"
	"github.com/spf13/cobra"
)

func main() {

	rootCmd.PersistentFlags().StringP("author", "a", "YOUR NAME", "author name for copyright attribution")

	// var rootCmd = &cobra.Command{Use: "app"}
	rootCmd.AddCommand(cmdEcho)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

var rootCmd = &cobra.Command{
	Use:   "kunst",
	Short: "kunst is the backend for wolfgang-ihle.de",
	Long: `
	A Fast and Flexible Static Site Generator built with
	love by spf13 and friends in Go.
	Complete documentation is available at http://hugo.spf13.com
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("asdf")
	},
}

var cmdEcho = &cobra.Command{
	Use:   "echo [string to echo]",
	Short: "Echo anything to the screen",
	Long: `echo is for echoing anything back.
Echo works a lot like print, except it has a child command.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Echo: " + strings.Join(args, " "))
	},
}

func kunst() {
	oap, err := hidrive.NewOauthProvider(hidrive.AppConfig{ClientID: ClientID, ClientSecret: ClientSecret})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	hdclient := hidrive.NewClient(oap, "/users/matt.ihle/wolfgang-ihle")
	hd := hidrive.NewDrive(oap, hidrive.PrefixPath("/users/matt.ihle/wolfgang-ihle"))

	srv := httpsrvr.NewServer("", 8000, false, false, nil)

	srv.Register("/hidrive", hidrive.HiDriveHandler(hd))
	srv.Register("/api", kunst.KunstHandler(flags.Database, hdclient))
	srv.Register("/media", hd)
	srv.Register("/api/signin", auth.SigninHandler(map[string]string{"matt": "$2a$14$4zu/jv7JO377BBg0k6upZ.Ul0jqO9enCBhHlAUyoKJrcySb8JMzW2", "wolfgang": "$2a$14$KWkdJOJLa4FkKHyXZ9xFceutb8qqkQ0V2Ue1.Ce9Rn0OD69.tDHHC"}))

	srv.Register("/api/auth", oap)
	srv.Register("/hidrive-token-callback", http.HandlerFunc(oap.HandleAuthorizeCallback))

	srv.ListenAndServe()
}
