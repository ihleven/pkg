package httpsrvr

import (
	"net/http"
	"path"
	"strings"
)

func NewDispatcher(handler http.Handler, name string) *dispatcher {
	if handler == nil {
		handler = http.NotFoundHandler()
	}
	return &dispatcher{name: name, handler: handler, children: make(map[string]*dispatcher)}
}

type dispatcher struct {
	name     string
	handler  http.Handler
	children map[string]*dispatcher
	preserve bool
}

func (r *dispatcher) PreservePath(preserve bool) *dispatcher {

	r.preserve = preserve
	return r
}

func (r *dispatcher) Name(name string) *dispatcher {

	r.name = name
	return r
}

func (r *dispatcher) Register(path string, handler http.Handler) *dispatcher {

	head, tail := shiftPath(path)

	switch {
	case path == "/":
		// root level
		r.handler = handler
		return r
	case tail == "/":
		// child route
		r.children[head] = NewDispatcher(handler, path[1:]) // {children: make(map[string]*dispatcher), handler: handler}
		return r.children[head]

	default:
		// nested child route
		if _, ok := r.children[head]; !ok {
			r.children[head] = NewDispatcher(r.handler, path) // r.handler -> notfound handler
		}
		return r.children[head].Register(tail, handler)
	}
}

func (d *dispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.handler.ServeHTTP(w, r)
}

func (d *dispatcher) GetDispatcher(route string) (*dispatcher, string) {

	head, tail := shiftPath(route)

	if disp, ok := d.children[head]; ok {
		return disp.GetDispatcher(tail)
	}
	return d, route
}

func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}
