package httpsrvr

import "time"

type logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Fatal(err error, format string, args ...interface{})
}
type accesslogger interface {
	Access(reqNum uint64, reqID string, start time.Time, addr, user, method, uri, proto string, status, size int, duration time.Duration, referer, agent string)
}

// https://pmihaylov.com/go-structured-logs/
