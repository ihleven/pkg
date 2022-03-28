package errors

import (
	"encoding/json"
	"fmt"
	"strconv"

	"go.uber.org/zap/zapcore"
)

func (e *Error) Format(f fmt.State, verb rune) {

	switch verb {
	case 's':
		var str string
		if f.Flag('#') || f.Flag('+') {
			for _, a := range e.Frames {
				str += a.Message + ": "
			}
		}
		str += e.cause.Error()
		f.Write([]byte(str))

	case 'v':
		var bytes []byte
		var err error
		if f.Flag('#') || f.Flag('+') {
			bytes, err = json.MarshalIndent(e, "", "    ")
		} else {
			bytes, err = json.Marshal(e)
		}
		if err != nil {
			bytes = []byte(err.Error())
		}
		f.Write(bytes)
	}
}

// MarshalJSON implements stdlib json interface
// This is used for formatting the error with the 'v' verb
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
	return json.MarshalIndent(foo, "", "    ")
}

// MarshalLogObject implements the interface used by zap error logging,
// used for appending potential errors to the request log.
func (a *Error) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("cause", a.cause.Error())
	enc.AddArray("ann", annotations(a.Frames))
	return nil
}

type annotations []Annotation

func (aa annotations) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	var err error
	for i := range aa {
		arr.AppendObject(aa[i])
	}
	return err
}

func (a Annotation) MarshalLogObject(enc zapcore.ObjectEncoder) error {

	if a.code != 0 {
		enc.AddInt("code", a.code)
	}
	enc.AddString(a.Message, a.File+":"+strconv.Itoa(a.Line)+" "+a.Function+"()")
	// enc.AddString("msg", a.Message)
	return nil
}
