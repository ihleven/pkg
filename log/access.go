// Package log wraps logging and is loosely oriented on https://dave.cheney.net/2015/11/05/lets-talk-about-logging
package log

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
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

var timeFormat = "02/Jan/2006:15:04:05 -0700"

type AccessLogger struct {
	Format string
}

func (l AccessLogger) Access(reqNum uint64, reqID string, start time.Time, remoteAddr, username, method, uri, proto string, status, size int, duration time.Duration, referer, agent string) {

	return
	// username := "-"

	// if req.URL.User != nil {
	// 	if name := req.URL.User.Username(); name != "" {
	// 		username = name
	// 	}
	// }

	switch l.Format {
	case "CombineLoggerType":
		fmt.Fprintln(os.Stdout, strings.Join([]string{
			" === access logger === ",
			remoteAddr,
			"-",
			reqID, //strconv.Itoa(int(reqNum)),
			"[" + start.Format(timeFormat) + "]",
			`"` + method,
			uri,
			proto + `"`,
			strconv.Itoa(status),
			strconv.Itoa(size),
			duration.String(),
			`"` + referer + `"`,
			`"` + agent + `"`,
		}, " "))

	}
}

func parseResponseTime(start time.Time) string {
	return fmt.Sprintf("%.3f ms", time.Now().Sub(start).Seconds()/1e6)
}
