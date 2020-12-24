// Package log wraps logging and is loosely oriented on https://dave.cheney.net/2015/11/05/lets-talk-about-logging
package log

import (
	"net/http"
	"net/url"
	"time"

	"github.com/fatih/color"
)

func Access(url url.URL, statusCode int, duration time.Duration, format string, args ...interface{}) {
	color.Cyan(" * %s?%s   Status: %v, took: %v => "+format+"\n", url.Path, url.RawQuery, statusCode, duration)

}

type RequestLogger interface {
	NewRequest(int64, time.Time, *http.Request) error
	Done(int64, time.Time, string) error
}

// var Logger RequestLogger

func NewRequest(counter int64, ts time.Time, r *http.Request) {
	// err := Logger.NewRequest(counter, ts, r)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

func Done(counter int64, rtype string) {
	// elapsed := time.Now()
	// color.Cyan(" *  %7d => %v \n", counter, elapsed)
	// err := Logger.Done(counter, elapsed, rtype)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
