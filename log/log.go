package log

import (
	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Debug logs stuff that the user doesn’t care about but developers do.
func Debugf(message string, args ...interface{}) {
	color.Cyan(message+"\n", args...)
}

// Info logs stuff that you tell the user
func Infof(format string, args ...interface{}) {
	// color.Green(format+"\n", args...)
	globalSugar.Infof(format, args...)
}

// Fatal -> one message, that exits the program.
func Errorf(err error, format string, args ...interface{}) {
	globalSugar.Errorf(format, args)
}

func Debug(msg string, fields ...Field) {
	globalPlain.Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	globalPlain.Info(msg, fields...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	globalSugar.Infow(msg, keysAndValues...)
}

type Field = zapcore.Field

var Str = zap.String
var Int = zap.Int

// var Err = zap.Any

func Err(e error) Field {
	return zap.Any("error", e)
}

// func Str(key string, val string) Field {
// 	return zap.String(key, val)
// }

// Info logs stuff that you tell the user
// func Infof(format string, args ...interface{}) {
// 	global.Sugar().Infof(format, args...)
// }
