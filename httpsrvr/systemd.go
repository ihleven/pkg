// +build linux

package httpsrvr

import (
	"log"
	"net/http"

	"github.com/coreos/go-systemd/activation"
	"github.com/coreos/go-systemd/daemon"
	"github.com/ihleven/errors"
)

func init() {

	listenAndServeSystemD = func(s *http.Server) error {
		log.Println("Serving on Linux with SystemD ...")

		listeners, err := activation.Listeners()
		if err != nil {
			return errors.Wrap(err, "cannot retrieve systemd listeners")
		}

		if len(listeners) == 0 {
			return errors.New("cannot retrieve systemd listeners")
		}

		// Die Notification für Systemd
		// soll bewusst vor "Serve" stehen!
		// siehe https://vincent.bernat.ch/en/blog/2017-systemd-golang
		// und https://vincent.bernat.ch/en/blog/2018-systemd-golang-socket-activation
		daemon.SdNotify(false, "READY=1")

		return s.Serve(listeners[0])
	}
}
