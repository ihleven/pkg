package httpsrvr

import (
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Error(err error, format string, args ...interface{})
	// Fatal(err error, format string, args ...interface{})
}

// https://pmihaylov.com/go-structured-logs/

func NewZapRequestLogger2() *zap.Logger {
	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "request",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey: "time",
			// EncodeTime: zapcore.ISO8601TimeEncoder,

			// CallerKey:    "caller",
			// EncodeCaller: zapcore.ShortCallerEncoder,

			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			// EncodeDuration: zapcore.MillisDurationEncoder,
		},
	}

	l, err := cfg.Build()
	if err != nil {
		return nil
	}

	return l
}
func NewZapRequestLogger(f *os.File) *zap.Logger {
	// highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
	// 	return lvl >= zapcore.ErrorLevel
	// })
	// lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
	// 	return lvl < zapcore.ErrorLevel
	// })
	// topicDebugging := zapcore.AddSync(ioutil.Discard)
	// topicErrors := zapcore.AddSync(ioutil.Discard)
	// consoleDebugging := zapcore.Lock(os.Stdout)
	// consoleErrors := zapcore.Lock(os.Stderr)

	// jsonEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	// consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	pe := zap.NewProductionEncoderConfig()
	pe.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewJSONEncoder(pe)
	consoleEncoder := zapcore.NewConsoleEncoder(pe)

	level := zap.InfoLevel

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
		zap.String("endpoint", r.URL.Path),
		zap.Any("query", r.URL.Query()),
		zap.Duration("duration", duration),
		zap.Uint64("size", rw.count),
		zap.Int("status", rw.statusCode),
	}
	if rw.err != nil {
		// fields = append(fields, zap.Error(fmt.Errorf("Inbound request failed with status %d", rw.statusCode)))
		fields = append(fields, zap.Error(rw.err))
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
