package regextable

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/ihleven/errors"
)

func newRoute(method, pattern string, handler paramhandler) route {
	return route{method, regexp.MustCompile("^" + pattern + "$"), nil, handler, nil, nil}
}

func NewRouter() *router {

	errorhandler := func(w http.ResponseWriter, r *http.Request, err error) {
		code := errors.Code(err)
		if code == 65535 {
			code = 500
		}
		w.WriteHeader(code)
		fmt.Fprintf(w, " -> %#v", err)
		fmt.Printf(" -> %#v\n", err)

	}

	return &router{
		notFoundHandler: http.NotFound,
		errorHandler:    errorhandler,
	}
}

type router struct {
	routes          []route
	errorHandler    func(http.ResponseWriter, *http.Request, error)
	notFoundHandler http.HandlerFunc
}

type paramhandler func(http.ResponseWriter, *http.Request, map[string]interface{})
type paramerrorhandler func(http.ResponseWriter, *http.Request, map[string]interface{}) error
type ehandler func(http.ResponseWriter, *http.Request) error

type route struct {
	method       string
	regex        *regexp.Regexp
	paramtypes   map[string]string
	handler      paramhandler //func(http.ResponseWriter, *http.Request, map[string]string)
	errorhandler paramerrorhandler
	ehandler     ehandler
}

func (r *router) Get(pattern string, handler interface{}) {
	r.Register(http.MethodGet, pattern, handler)
}

func (r *router) Post(pattern string, handler interface{}) {
	r.Register(http.MethodPost, pattern, handler)
}

func (r *router) Put(pattern string, handler interface{}) {
	r.Register(http.MethodPut, pattern, handler)
}

func (r *router) Patch(pattern string, handler interface{}) {
	r.Register(http.MethodPatch, pattern, handler)
}

func (r *router) Delete(pattern string, handler interface{}) {
	r.Register(http.MethodDelete, pattern, handler)
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
		r.routes = append(r.routes, route{method, routeRE, types, nil, ht, nil})

	case func(http.ResponseWriter, *http.Request, map[string]interface{}):
		r.routes = append(r.routes, route{method, routeRE, types, ht, nil, nil})

	case func(http.ResponseWriter, *http.Request) error:
		r.routes = append(r.routes, route{method, routeRE, types, nil, nil, ht})
	default:
		fmt.Println("DEFAULT")
	}

}
func (rou *router) Dispatch(r *http.Request) (*route, map[string]interface{}, int) {
	var allow []string

	for _, route := range rou.routes {
		matches := route.regex.FindStringSubmatch(r.URL.Path)

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
							return nil, nil, 500
						}
						params[name] = intval
					default:
						params[name] = matches[i]
					}
				}
			}
			return &route, params, 200
		}
	}

	if len(allow) > 0 {
		// w.Header().Set("Allow", strings.Join(allow, ", "))
		return nil, nil, 405
	}
	return nil, nil, 404
}

func (rou *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var allow []string

	for _, route := range rou.routes {
		matches := route.regex.FindStringSubmatch(r.URL.Path)

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
			} else if route.errorhandler != nil {
				err := route.errorhandler(w, r, params)
				if err != nil {
					rou.errorHandler(w, r, err)
				}
			} else {
				err := route.ehandler(w, r)
				if err != nil {
					fmt.Println("err:", err)
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
