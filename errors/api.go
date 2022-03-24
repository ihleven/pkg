package errors

import "fmt"

func New(msg string, vals ...interface{}) error {

	return &Error{
		cause:      fmt.Errorf(msg, vals...),
		stacktrace: trace(2),
	}
}

// Wrap adds Annotation to existing error.
func Wrap(err error, msg string, vals ...interface{}) error {
	if err == nil {
		return nil
	}

	e := create(err, msg, vals...)
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
