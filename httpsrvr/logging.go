package httpsrvr

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// type logger interface {
// 	Debug(format string, args ...interface{})
// 	Info(format string, args ...interface{})
// 	Error(err error, format string, args ...interface{})
// 	// Fatal(err error, format string, args ...interface{})
// }

// https://pmihaylov.com/go-structured-logs/

func NewZapRequestLogger(f *os.File) *zap.Logger {

	encoderConfig := zapcore.EncoderConfig{
		MessageKey: "request",

		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,

		TimeKey: "time",
		// EncodeTime: zapcore.ISO8601TimeEncoder,

		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,

		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		// EncodeDuration: zapcore.MillisDurationEncoder,
	}
	encoderConfig = zap.NewProductionEncoderConfig()
	encoderConfig = zap.NewDevelopmentEncoderConfig()

	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	level := zap.InfoLevel // zap.NewAtomicLevelAt(zapcore.InfoLevel)

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(f), level),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
	)

	logger := zap.New(core)
	defer logger.Sync()

	return logger
}

func (s *httpServer) LogRequest(r *http.Request, id string, reqnum uint64, start time.Time, rw *ResponseWriter, name string) {

	duration := time.Since(start)

	color.Green("request %d: %s %s => %d (%d bytes, %v)\n", reqnum, id, r.URL.Path, rw.statusCode, rw.Count(), duration)

	fields := []zap.Field{

		zap.Time("ts", start),
		zap.String("instance", s.instance),
		zap.Uint64("num", reqnum),
		zap.String("id", id),
		zap.String("addr", r.RemoteAddr),
		zap.String("referer", r.Referer()),
		zap.String("proto", r.Proto),
		zap.String("handler", name),
		zap.String("method", r.Method),
		zap.String("handlerpath", r.URL.Path),
		zap.String("authkey", rw.Authkey),
		zap.Any("query", r.URL.Query()),
		zap.Duration("duration", duration),
		zap.Uint64("size", rw.count),
		zap.Int("status", rw.statusCode),
	}
	if rw.err != nil {
		// fields = append(fields, zap.Error(fmt.Errorf("Inbound request failed with status %d", rw.statusCode)))
		// fields = append(fields, zap.Error(rw.err))
		fields = append(fields, zap.Any("err", rw.err))
	}
	if rw.RuntimeStack != nil {
		fields = append(fields, zap.ByteString("runtimeStack", rw.RuntimeStack))
	}
	if rw.statusCode >= 400 {
		s.requestLogger.Error("", fields...)
	} else {
		s.requestLogger.Info("", fields...)
	}

	// color.Green("request %d: %s %s => %d (%d bytes, %v)\n\n", reqnum, reqid, r.URL.Path, rw.statusCode, rw.Count(), duration)

	// defer func(start time.Time, reqnum uint64, reqid string, name string) {
	// 	fmt.Println("defer ==========================", start, time.Now())
	// 	// s.logger.Request(r, reqid, reqnum, start, rw.statusCode, rw.Count(), name, rw.err)
	// color.Green("req %d: %s %s => %d (%d bytes, %v)\n", reqnum, reqid, r.URL.Path, rw.statusCode, rw.Count(), time.Since(start))

	// 	s.LogRequest(r, reqid, reqnum, start, rw, name)
	// }(start, reqnum, reqid, route.name)
}

var (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
)

// Colorize wraps a given message in a given color.
//
// Example:
//     println(color.Colorize(color.Red, "This is red"))
func Colorize(color, s string) string {
	return color + s + Reset
}

var zerologger zerolog.Logger

func init() {
	zerologger = zerolog.New(os.Stderr).With().Logger()
}
func (s *httpServer) zerologRequest(r *http.Request, id string, reqnum uint64, start time.Time, rw *ResponseWriter, name string) {
	duration := time.Since(start)

	// fmt.Printf("%srequest %d: %s %s => %d (%d bytes, %v)", Purple, reqnum, id, r.URL.Path, rw.statusCode, rw.Count(), duration)
	fmt.Println(Purple)
	zerologger.Log().Uint(
		"reqnum", uint(reqnum),
	).Str(
		"name", name,
	).Str(
		"reqid", id,
	).Str(
		"method", r.Method,
	).Str(
		"path", r.URL.Path,
	).Msgf("request %d: %s %s => %d (%d bytes, %v)", reqnum, id, r.URL.Path, rw.statusCode, rw.Count(), duration)
	fmt.Println(Reset)

}
