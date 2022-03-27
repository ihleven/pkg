package errors

import (
	"fmt"
	"runtime"
)

func create(cause error, code int, msg string, vals ...interface{}) *Error {

	e, ok := cause.(*Error)
	if !ok {
		e = &Error{
			cause:      cause,
			stacktrace: trace(3),
		}
	}

	e.createAnnotation(code, msg, vals...)

	return e
}

type Error struct {
	Frames        []Annotation
	cause         error
	stacktrace    []StackFrame
	pkgerrorstack *stack
}

func (e *Error) Error() string {
	return e.cause.Error()
}

func (e *Error) Cause() error {
	return e.cause
}

func (e *Error) Code() int {
	for _, a := range e.Frames {
		if a.code != 0 {
			return a.code
		}
	}
	return 500
}

// Unwrap returns the original error.
func (e *Error) Unwrap() error {
	return e.cause
}

// Frame is a single step in stack trace.
type StackFrame struct {
	Func string
	Line int
	Path string
}

func (f StackFrame) String() string {
	return fmt.Sprintf("%s:%d %s()", f.Path, f.Line, f.Func)
}

func trace(skip int) []StackFrame {
	frames := make([]StackFrame, 0, 64)
	for {
		pc, path, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		frame := StackFrame{
			Func: fn.Name(),
			Line: line,
			Path: path,
		}
		frames = append(frames, frame)
		skip++
	}
	return frames
}
