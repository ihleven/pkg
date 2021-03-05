package regextable

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func newRoute(method, pattern string, handler paramhandler) route {
	return route{method, regexp.MustCompile("^" + pattern + "$"), nil, handler, nil}
}

func NewRouter() *router {

	return &router{notFoundHandler: http.NotFound}
}

type router struct {
	routes          []route
	errorHandler    func(http.ResponseWriter, *http.Request, error)
	notFoundHandler http.HandlerFunc
}

type paramhandler func(http.ResponseWriter, *http.Request, map[string]interface{})
type paramerrorhandler func(http.ResponseWriter, *http.Request, map[string]interface{}) error

type route struct {
	method       string
	regex        *regexp.Regexp
	paramtypes   map[string]string
	handler      paramhandler //func(http.ResponseWriter, *http.Request, map[string]string)
	errorhandler paramerrorhandler
}

func (r *router) Register(method, pattern string, handler interface{}) {

	var pathelems []string
	re := regexp.MustCompile(`^{([^/:]+)(:([^/]*))?}$`)

	types := make(map[string]string)

	for _, subpattern := range strings.Split(pattern, "/") {
		elem := subpattern
		if matches := re.FindStringSubmatch(subpattern); len(matches) > 0 {
			n, t := matches[1], matches[3]
			if t == "" {
				t = "string"
			}
			types[n] = t
			switch t {
			case "int":
				elem = fmt.Sprintf("(?P<%s>[0-9]+)", n)
			default:
				elem = fmt.Sprintf("(?P<%s>[^/]+)", n)
			}
		}
		pathelems = append(pathelems, elem)
	}
	routeRE := regexp.MustCompile("^" + strings.Join(pathelems, "/") + "$")
	switch ht := handler.(type) {
	case func(http.ResponseWriter, *http.Request, map[string]interface{}) error:

		r.routes = append(r.routes, route{method, routeRE, types, nil, ht})
	case func(http.ResponseWriter, *http.Request, map[string]interface{}):

		r.routes = append(r.routes, route{method, routeRE, types, ht, nil})

	}

}

func (rou *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var allow []string

	for _, route := range rou.routes {
		matches := route.regex.FindStringSubmatch(r.URL.Path)
		fmt.Println("matches:", matches)
		if len(matches) > 0 {

			if !strings.Contains(route.method, r.Method) { // r.Method != route.method {
				allow = append(allow, route.method)
				continue
			}

			params := make(map[string]interface{})
			for i, name := range route.regex.SubexpNames() {

				if i != 0 && name != "" {
					switch route.paramtypes[name] {
					case "int":
						intval, err := strconv.Atoi(matches[i])
						if err != nil {
							rou.errorHandler(w, r, err)
							return
						}
						params[name] = intval
					default:
						params[name] = matches[i]
					}
				}
			}

			if route.handler != nil {
				// ctx := context.WithValue(r.Context(), ctxKey{}, matches[1:])
				route.handler(w, r, params) // .WithContext(ctx))
			} else {
				err := route.errorhandler(w, r, params)
				if err != nil {
					rou.errorHandler(w, r, err)
				}
			}
			return
		}
	}
	if len(allow) > 0 {
		w.Header().Set("Allow", strings.Join(allow, ", "))
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rou.notFoundHandler(w, r)
}
