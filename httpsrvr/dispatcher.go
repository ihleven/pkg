package httpsrvr

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

func NewDispatcher(handler http.Handler) *dispatcher {
	if handler == nil {
		handler = http.NotFoundHandler()
	}
	return &dispatcher{children: make(map[string]*dispatcher), handler: handler}
}

type dispatcher struct {
	handler  http.Handler
	children map[string]*dispatcher
	preserve bool
}

func (r *dispatcher) PreservePath(preserve bool) *dispatcher {

	r.preserve = preserve
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
		r.children[head] = NewDispatcher(handler) // {children: make(map[string]*dispatcher), handler: handler}
		return r.children[head]

	default:
		// nested child route
		if _, ok := r.children[head]; !ok {
			r.children[head] = NewDispatcher(r.handler) // &dispatcher{children: make(map[string]*dispatcher), handler: r.handler}
		}
		return r.children[head].Register(tail, handler)
	}
}

func (d *dispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	disp, tail := d.getDispatcher(r.URL.Path)
	if !disp.preserve {
		fmt.Println("preserve:", disp)
		r.URL.Path = tail
	}
	disp.handler.ServeHTTP(w, r)
}
func (d *dispatcher) getDispatcher(route string) (*dispatcher, string) {
	head, tail := shiftPath(route)
	// fmt.Printf("GetHandler => %q - %q     %v\n", head, tail, d)
	if disp, ok := d.children[head]; ok {
		return disp.getDispatcher(tail)
	}
	// fmt.Printf("GetHandler => %v\n", d)
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
