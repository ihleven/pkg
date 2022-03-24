package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
)

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

// Unwrap returns the original error.
func (e *Error) Unwrap() error {
	return e.cause
}

func (e *Error) MarshalJSON() ([]byte, error) {

	var frames []string
	for _, f := range e.stacktrace {
		frames = append(frames, f.String())
	}
	var frames2 []string
	for _, f := range *e.pkgerrorstack {
		frame := Frame(f)
		frames2 = append(frames, fmt.Sprintf("%v", frame))
	}
	foo := struct {
		Frames      []Annotation `json:"annotations,omitempty"`
		Cause       string       `json:"cause,omitempty"`
		Stacktrace  []string     `json:"stacktrace,omitempty"`
		Stacktrace2 []string     `json:"stacktrace2,omitempty"`
	}{Cause: e.cause.Error(), Frames: e.Frames, Stacktrace: frames, Stacktrace2: frames2}
	// fmt.Printf("stack: %#v\n", e.pkgerrorstack)
	return json.Marshal(foo)
}

func create(cause error, msg string, vals ...interface{}) *Error {

	e, ok := cause.(*Error)
	if !ok {
		e = &Error{
			cause:      cause,
			stacktrace: trace(3),
		}
	}

	e.createAnnotation(msg, vals...)

	return e
}

//////////////////////////////////////////////////////////////////////////////////

// func (e *Error) Format(st fmt.State, verb rune) {
// 	switch verb {
// 	case 'v':
// 		switch {
// 		case st.Flag('+'):
// 			fallthrough

// 		case st.Flag('#'):
// 			for _, frame := range e.frames {

// 				fmt.Fprintf(st, "\n%+v", frame)
// 			}
// 		}
// 	}
// }

// Error returns error message.
// func (e *errorData) Error() string {
// 	return e.error.Error()
// }

// Frame is a single step in stack trace.
type StackFrame struct {
	// Func contains a function name.
	Func string
	// Line contains a line number.
	Line int
	// Path contains a file path.
	Path string
}

// String formats Frame to string.
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
