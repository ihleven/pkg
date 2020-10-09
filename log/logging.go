// Package log is a very simple logger
package log

import (
	"os"

	"github.com/fatih/color"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
)

func NewStdoutLogger(level LogLevel) *Logger {
	return &Logger{Level: level}
}

type Logger struct {
	Level LogLevel
}

// Debug logs stuff that the user doesn’t care about but developers do.
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.Level == DEBUG {
		color.White(" ---- "+format+"\n", args...)
	}
}

// Info logs stuff that you tell the user
func (l *Logger) Info(format string, args ...interface{}) {
	color.Cyan(" --- "+format+"\n", args...)
}

// Fatal -> one message, that exits the program.
func (l *Logger) Fatal(err error, format string, args ...interface{}) {
	// stdlog.Fatal(v)

	color.Red(" --- Fatal error, exiting with code 1: "+format+"\n", args...)
	if err != nil {
		color.Red("%+v", err)
	}
	os.Exit(1)
}
