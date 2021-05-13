package main

import (
	"fmt"
	"os"

	"github.com/ihleven/pkg/auth"
	"github.com/ihleven/pkg/hidrive"
	"github.com/ihleven/pkg/httpsrvr"
	"github.com/spf13/cobra"
)

var (
	version      = ""
	clientID     = "a7ff922a897cfde20473f1b8d01b42a9"
	clientSecret = "e787251e8d6817fdaf7f5c6d53dc8a0d"
)

func main() {

	cmd.PersistentFlags().StringP("author", "a", "YOUR NAME", "author name for copyright attribution")
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
	}

}

var cmd = &cobra.Command{
	Use:   "hidrive",
	RunE:  run,
	Short: "hidrive file server",
	// Long: `
	// `,
}

func run(cmd *cobra.Command, args []string) error {

	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	authentication := auth.NewAuthentication()

	wi := hidrive.NewDrive(clientID, clientSecret, hidrive.Prefix("/wolfgang-ihle"))

	srvr := httpsrvr.NewServer(8000, true)

	srvr.Register("/hidrive", wi)
	srvr.Register("/home", hidrive.NewDrive(clientID, clientSecret, hidrive.FromHome()))
	// srvr.Register("/hidrive/auth", authMngr)
	srvr.Register("/login", authentication.LoginHandler).PreservePath()
	srvr.Register("/logout", authentication.LogoutHandler)

	srvr.Run()

	return nil
}
