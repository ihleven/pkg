package log

import (
	"net/http"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (l *Logger) LogRequest(num uint64, id string, start time.Time, addr, referer string, method, path, proto string, statuscode int, respSize uint64, respDuration time.Duration, name string) {

	l.zapLogger.Info("This is an INFO message with fields",
		zap.String("request", id),
		zap.Uint64("num", num),
		zap.Time("start", start),
		zap.String("addr", addr),
		zap.String("referer", referer),
		zap.String("proto", proto),
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("statuscode", statuscode),
		zap.Uint64("size", respSize),
		zap.Duration("duration", respDuration),
		zap.String("name", name),
	)

}

func (l *Logger) Request(r *http.Request, id string, reqnum uint64, start time.Time, statusCode int, respSize uint64, name string, err error) {
	var reqid string
	if id := r.Context().Value("reqid"); id != nil {
		reqid = id.(string)
	}

	duration := time.Since(start)

	l.zapLogger.Info("",
		zap.Time("start", start),
		zap.Uint64("reqnum", reqnum),
		zap.String("reqid", reqid),
		zap.String("addr", r.RemoteAddr),
		zap.String("referer", r.Referer()),
		zap.String("proto", r.Proto),
		zap.String("handler", name),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Duration("duration", duration),
		zap.Uint64("size", respSize),
		zap.Int("status", statusCode),
		zap.Error(err),
	)
	color.Green("request %d: %s %s => %d (%d bytes, %v)\n\n", reqnum, reqid, r.URL.Path, statusCode, respSize, duration)

}

type RequestLogger zap.Logger

func NewRequestLogger() *RequestLogger {
	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig:    zapcore.EncoderConfig{
			// MessageKey: "message",

			// LevelKey:    "level",
			// EncodeLevel: zapcore.CapitalLevelEncoder,

			// TimeKey:    "time",
			// EncodeTime: zapcore.ISO8601TimeEncoder,

			// CallerKey:    "caller",
			// EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	l, err := cfg.Build()
	if err != nil {
		return nil
	}

	return (*RequestLogger)(l)
}

func (l *RequestLogger) LogRequest(r *http.Request, id string, reqnum uint64, start time.Time, statusCode int, respSize uint64, name string) {
	var reqid string
	if id := r.Context().Value("reqid"); id != nil {
		reqid = id.(string)
	}

	duration := time.Since(start)

	(*zap.Logger)(l).Info("",
		zap.Time("start", start),
		zap.Uint64("num", reqnum),
		zap.String("id", reqid),
		// zap.String("addr", r.RemoteAddr),
		// zap.String("referer", r.Referer()),
		// zap.String("proto", r.Proto),
		zap.String("handler", name),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Duration("duration", duration),
		zap.Uint64("size", respSize),
		zap.Int("status", statusCode),
	)
	color.Green("request %d: %s %s => %d (%d bytes, %v)\n\n", reqnum, reqid, r.URL.Path, statusCode, respSize, duration)

}
