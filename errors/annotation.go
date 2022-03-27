package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
)

type Annotation struct {
	Message  string
	code     int
	File     string
	Function string
	Line     int
}

func (st *Annotation) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Code    int    `json:"code,omitempty"`
		Message string `json:"msg,omitempty"`
		At      string `json:"at,omitempty"`
	}{
		Code:    st.code,
		Message: st.Message,
		At:      fmt.Sprintf("%s:%d (%s)", st.File, st.Line, st.Function),
	})
}

func (e *Error) createAnnotation(code int, msg string, vals ...interface{}) {

	st := Annotation{
		code:    code,
		Message: fmt.Sprintf(msg, vals...),
	}

	pc, file, line, ok := runtime.Caller(3)
	if ok {
		st.File, st.Line = file, line

		f := runtime.FuncForPC(pc)
		if f != nil {
			st.Function = shortFuncName(f)
		}

	}

	e.Frames = append([]Annotation{st}, e.Frames...)
}

/* "FuncName" or "Receiver.MethodName" */
func shortFuncName(f *runtime.Func) string {
	// f.Name() is like one of these:
	// - "github.com/palantir/shield/package.FuncName"
	// - "github.com/palantir/shield/package.Receiver.MethodName"
	// - "github.com/palantir/shield/package.(*PtrReceiver).MethodName"
	longName := f.Name()

	withoutPath := longName[strings.LastIndex(longName, "/")+1:]
	// withoutPackage := withoutPath[strings.Index(withoutPath, ".")+1:]

	shortName := withoutPath
	// shortName = strings.Replace(shortName, "(", "", 1)
	// shortName = strings.Replace(shortName, "*", "", 1)
	// shortName = strings.Replace(shortName, ")", "", 1)

	return shortName
}
