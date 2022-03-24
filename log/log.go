package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var global *zap.Logger

func init() {
	// global, _ = zap.NewProduction(zap.AddCallerSkip(1))
	// defer logger.Sync()

	cfg := zap.Config{
		Encoding:         "console",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	global, _ = cfg.Build()
}

func Debug(msg string, fields ...Field) {
	global.Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	global.Info(msg, fields...)
}

type Field = zapcore.Field

var Str = zap.String
var Int = zap.Int

// func Int(key string, val int) Field {
// 	return zap.Int(key, val)
// }

// func Str(key string, val string) Field {
// 	return zap.String(key, val)
// }

// Info logs stuff that you tell the user
func Infof(format string, args ...interface{}) {
	global.Sugar().Infof(format, args...)
}

// Fatal -> one message, that exits the program.
func Errorf(err error, format string, args ...interface{}) {
	global.Sugar().Infof(format, args)
}
