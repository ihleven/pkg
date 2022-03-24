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

// func (st *Annotation) Error() string {
// 	return fmt.Sprint(st)
// }

func (st *Annotation) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Message string `json:"msg,omitempty"`
		Code    string `json:"code,omitempty"`
		At      string `json:"at,omitempty"`
	}{
		Message: st.Message,
		At:      fmt.Sprintf("%s:%d (%s)", st.File, st.Line, st.Function),
	})
	// return json.Marshal(map[string]string{st.Message: fmt.Sprintf("%s:%d (%s)", st.File, st.Line, st.Function)})
}

func (e *Error) createAnnotation(msg string, vals ...interface{}) {

	st := Annotation{
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

/*
DefaultFormat defines the behavior of err.Error() when called on a stacktrace,
as well as the default behavior of the "%v", "%s" and "%q" formatting
specifiers. By default, all of these produce a full stacktrace including line
number information. To have them produce a condensed single-line output, set
this value to stacktrace.FormatBrief.
The formatting specifier "%+s" can be used to force a full stacktrace regardless
of the value of DefaultFormat. Similarly, the formatting specifier "%#s" can be
used to force a brief output.
*/
var DefaultFormat = FormatFull

// Format is the type of the two possible values of stacktrace.DefaultFormat.
type Format int

const (
	// FormatFull means format as a full stacktrace including line number information.
	FormatFull Format = iota
	// FormatBrief means Format on a single line without line number information.
	FormatBrief
)

// var _ fmt.Formatter = (*Annotation)(nil)

func (e *Error) Format(f fmt.State, c rune) {
	// fmt.Printf("+++++++++++++++++++++++%v++++++++++++++++++++++++++++++\n", e)

	bytes, err := json.Marshal(e)
	// fmt.Printf("++++++++++++++++++++++++++++%q+++++++++++++++++++++++++\n", bytes)
	if err != nil {
		f.Write([]byte(err.Error()))
		return
	}
	f.Write(bytes)
}

func (st *Annotation) Format2(f fmt.State, c rune) {
	var text string
	if f.Flag('+') && !f.Flag('#') && c == 's' { // "%+s"
		text = formatFull(st)
	} else if f.Flag('#') && !f.Flag('+') && c == 's' { // "%#s"
		text = formatBrief(st)
	} else {
		text = map[Format]func(*Annotation) string{
			FormatFull:  formatFull,
			FormatBrief: formatBrief,
		}[DefaultFormat](st)
	}

	formatString := "%"
	// keep the flags recognized by fmt package
	for _, flag := range "-+# 0" {
		if f.Flag(int(flag)) {
			formatString += string(flag)
		}
	}
	if width, has := f.Width(); has {
		formatString += fmt.Sprint(width)
	}
	if precision, has := f.Precision(); has {
		formatString += "."
		formatString += fmt.Sprint(precision)
	}
	formatString += string(c)
	fmt.Fprintf(f, formatString, text)
}

func formatFull(st *Annotation) string {
	var str string

	return str
}

func formatBrief(st *Annotation) string {
	var str string

	return str
}
