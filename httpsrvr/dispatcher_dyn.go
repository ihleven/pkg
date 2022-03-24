package httpsrvr

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/ihleven/pkg/log"
)

func NewShiftPathRouter(handler http.Handler, name string) *ShiftPathRoute {
	if handler == nil {
		handler = http.NotFoundHandler()
	}
	return &ShiftPathRoute{
		Name:    name,
		Handler: map[string]http.Handler{"": handler},
		PathMap: make(map[string]*ShiftPathRoute),
	}
}

type ShiftPathRoute struct {
	Name     string
	Handler  map[string]http.Handler
	PathMap  map[string]*ShiftPathRoute
	RegexMap map[*regexp.Regexp]*ShiftPathRoute
}

func (d *ShiftPathRoute) getRegexSubRoute(head string) (*ShiftPathRoute, bool) {

	isRegex := strings.HasPrefix(head, "(")
	if !isRegex {
		return nil, false
	}
	// regexp, err := regexp.Compile(head)

	// stringRegex := regexp.MustCompile("(?P<%s>[^/]+)")

	regex := regexp.MustCompile(head)

	// map anlegen sofern nocht nicht vorhanden
	if d.RegexMap == nil {
		d.RegexMap = make(map[*regexp.Regexp]*ShiftPathRoute)
	}

	// Eintrag anlegen
	if _, ok := d.RegexMap[regex]; !ok {
		d.RegexMap[regex] = NewShiftPathRouter(nil, head)
	}
	return d.RegexMap[regex], true
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
		return d.PathMap[head].Register(method, tail, handler)
	}
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
			fmt.Println("matches:", matches, regex.SubexpNames(), params)
			if matches == nil {
				continue
			}
			for i, name := range regex.SubexpNames() {

				if name != "" && matches[i] != "" {
					params[name] = matches[i]
				}
			}
			fmt.Println("matches:", matches, regex.SubexpNames(), params)

			return disp.Dispatch(tail, params)
		}
	}

	return d, route
}

func (d *ShiftPathRoute) ServeHTTPsrvr(rw *ResponseWriter, r *http.Request) {

	route, _ := d.Dispatch(r.URL.Path, rw.Params)

	if handler, ok := route.Handler[r.Method]; ok {
		handler.ServeHTTP(rw, r)
	} else {
		route.Handler[""].ServeHTTP(rw, r)
	}
}
