// Package log is a very simple logger
package log

import (
	"os"

	"github.com/fatih/color"
	"go.uber.org/zap"
)

type Level int

const (
	OFF Level = iota
	INFO
	DEBUG
)

type logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Fatal(err error, format string, args ...interface{})
}

func NewLogger(level Level) *Logger {
	//
	zl, _ := zap.NewDevelopment(zap.AddCallerSkip(1))
	defer zl.Sync()

	return &Logger{
		Level:     level,
		zapLogger: zl,
	}
}

type Logger struct {
	Level     Level
	zapLogger *zap.Logger
}

func (l *Logger) Close() {
	if l.zapLogger != nil {
		l.zapLogger.Sync()
	}
}

// Debug logs stuff that the user doesn’t care about but developers do.
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.Level == DEBUG {
		color.White(" ---- "+format+"\n\n", args...)
	}
}

// Info logs stuff that you tell the user
func (l *Logger) Info(format string, args ...interface{}) {
	color.Cyan(" --- "+format+"\n\n", args...)
}

// Fatal -> one message, that exits the program.
func (l *Logger) Error(err error, format string, args ...interface{}) {
	// stdlog.Fatal(v)

	color.Red(" --- Fatal error, exiting with code 1: "+format+"\n\n", args...)
	if err != nil {
		color.Red("%+v", err)
	}
	os.Exit(1)
}
