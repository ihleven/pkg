package errors

import "fmt"

func New(msg string, vals ...interface{}) error {

	return &Error{
		cause:      fmt.Errorf(msg, vals...),
		stacktrace: trace(2),
	}
}

func NewWithCode(code int, msg string, vals ...interface{}) error {

	err := &Error{
		cause:      fmt.Errorf(msg, vals...),
		stacktrace: trace(2),
	}
	err.createAnnotation(code, "")
	return err
}

// Wrap adds Annotation to existing error.
func Wrap(err error, msg string, vals ...interface{}) error {
	if err == nil {
		return nil
	}

	e := create(err, 0, msg, vals...)
	e.pkgerrorstack = callers()
	return e

}

// Wrap adds Annotation to existing error.
func WrapWithCode(err error, code int, msg string, vals ...interface{}) error {
	if err == nil {
		return nil
	}

	e := create(err, code, msg, vals...)
	e.pkgerrorstack = callers()
	return e

}

// If the error does not implement Cause, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}

func Code(err error) int {

	var code int

	type coder interface{ Code() int }

	if errWithCode, ok := err.(coder); ok {
		code = errWithCode.Code()
	}
	return code
}
