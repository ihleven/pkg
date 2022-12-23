package httpsrvr

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/ihleven/pkg/httpauth"
	"github.com/ihleven/pkg/log"
)

func NewShiftPathRouter(handler http.Handler, name string) *ShiftPathRoute {
	// if handler == nil {
	// 	handler = http.HandlerFunc(MethodNotAllowed)
	// }
	return &ShiftPathRoute{
		Name:     name,
		Handler:  map[string]http.Handler{"": handler},
		PathMap:  make(map[string]*ShiftPathRoute),
		RegexMap: make(map[*regexp.Regexp]*ShiftPathRoute),
	}
}

type ShiftPathRoute struct {
	Name     string
	Handler  map[string]http.Handler
	PathMap  map[string]*ShiftPathRoute
	RegexMap map[*regexp.Regexp]*ShiftPathRoute
	// OptionAuth       bool
	OptionParseRequestAuth AuthParser
}

func (d *ShiftPathRoute) getRegexSubRoute(head string) (*ShiftPathRoute, bool) {

	if splits := strings.SplitN(head, ":", 2); len(splits) == 2 {
		if splits[0] == "" {
			head = "(?P<" + splits[1] + ">^[0-9a-zA-Z]+$)"
		} else if splits[0] == "int" {
			head = "(?P<" + splits[1] + ">^[0-9]+$)"
		}
	} else {

		isRegex := strings.HasPrefix(head, "(")
		if !isRegex {
			return nil, false
		}
	}

	// stringRegex := regexp.MustCompile("(?P<%s>[^/]+)")

	regex := regexp.MustCompile(head)

	// map anlegen sofern nocht nicht vorhanden
	if d.RegexMap == nil {
		d.RegexMap = make(map[*regexp.Regexp]*ShiftPathRoute)
	}

	// Eintrag suchen und zurückgeben.
	// Das der Key ein regexptr ist müssen wir hier die Stringrepräsentation aller keys vergleichen
	for regexkey, route := range d.RegexMap {
		if regexkey.String() == regex.String() {
			return route, true
		}
	}
	// Nur Eintrag anlegen wenn nicht gefunden
	d.RegexMap[regex] = NewShiftPathRouter(nil, head)
	return d.RegexMap[regex], true
}

func (d *ShiftPathRoute) GET(path string, h interface{}) *ShiftPathRoute {
	return d.Register(http.MethodGet, path, h)
}
func (d *ShiftPathRoute) POS(path string, h interface{}) *ShiftPathRoute {
	return d.Register(http.MethodPost, path, h)
}
func (d *ShiftPathRoute) PUT(path string, h interface{}) *ShiftPathRoute {
	return d.Register(http.MethodPut, path, h)
}
func (d *ShiftPathRoute) PAT(path string, h interface{}) *ShiftPathRoute {
	return d.Register(http.MethodPatch, path, h)
}
func (d *ShiftPathRoute) DEL(path string, h interface{}) *ShiftPathRoute {
	return d.Register(http.MethodDelete, path, h)
}

func (d *ShiftPathRoute) Register(method, path string, h interface{}) *ShiftPathRoute {

	handler, err := ConvertHandlerType(h)
	if err != nil {
		log.Infof("Could not register route '%v': unknown handler type %T", path, h)
		os.Exit(1)
	}

	head, tail := ShiftPath(path)

	if head == "" {
		// root level
		d.Handler[method] = handler
		return d
	}

	if subEntry, isRegex := d.getRegexSubRoute(head); isRegex {
		if tail == "/" {
			subEntry.Handler[method] = handler
			return subEntry
		}
		return subEntry.Register(method, tail, handler)
	}

	if _, ok := d.PathMap[head]; !ok {
		d.PathMap[head] = NewShiftPathRouter(nil, head) // r.handler -> notfound handler
	}
	if tail == "/" {
		d.PathMap[head].Handler[method] = handler
		return d.PathMap[head]
	} else {
		return d.PathMap[head].Register(method, tail, handler).SetName(path)
	}
}

func (r *ShiftPathRoute) SetName(name string) *ShiftPathRoute {
	r.Name = name
	return r
}

func (r *ShiftPathRoute) Auth(auth *httpauth.Auth) *ShiftPathRoute {
	r.OptionParseRequestAuth = auth
	return r
}

//

func (d *ShiftPathRoute) Dispatch(route string, params map[string]string) (*ShiftPathRoute, string) {

	head, tail := ShiftPath(route)

	if disp, ok := d.PathMap[head]; ok {
		return disp.Dispatch(tail, params)
	}

	if head != "" && d.RegexMap != nil {
		for regex, disp := range d.RegexMap {

			matches := regex.FindStringSubmatch(head) // oder route ???
			if matches == nil {
				continue
			}
			for i, name := range regex.SubexpNames() {

				if name != "" && matches[i] != "" {
					params[name] = matches[i]
				}
			}

			return disp.Dispatch(tail, params)
		}
	}

	return d, route
}

func (d *ShiftPathRoute) GetHandler(method string) http.Handler {
	if handler, ok := d.Handler[method]; ok {
		return handler
	} else if handler, ok := d.Handler[""]; ok && handler != nil {
		return handler
	}

	return nil
}

func (d *ShiftPathRoute) ServeHTTPsrvr(rw *ResponseWriter, r *http.Request) {

	route, tail := d.Dispatch(r.URL.Path, rw.Params)
	handler := route.GetHandler(r.Method)
	fmt.Printf("ServeHTTPsrvr %#v - %s - %v - %s\n", route, tail, handler, r.URL.Path)
	if handler != nil {
		rw.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		rw.Header().Set("Access-Control-Allow-Credentials", "true")
		rw.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT")
		rw.Header().Set("Access-Control-Allow-Headers", "Content-Type,WithCredentials,Authorization,Cookie")

		handler.ServeHTTP(rw, r)
	} else if r.Method == "OPTIONS" {
		rw.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		rw.Header().Set("Access-Control-Allow-Credentials", "true")
		rw.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT")
		rw.Header().Set("Access-Control-Allow-Headers", "Content-Type,WithCredentials,Authorization,Cookie")

		rw.WriteHeader(200)
	} else {
		http.Error(rw, http.StatusText(http.StatusMethodNotAllowed)+" in ServeHTTPsrvr at dispatcher_dyn.go:194", http.StatusMethodNotAllowed)
		rw.RespondJSON(rw)
		rw.RespondJSON(r.Method)
		rw.RespondJSON(route)
		rw.RespondJSON(handler)
	}
}

// func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
// 	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
// }
