package httpsrvr

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

func NewDispatcher() *route {

	return &route{children: make(map[string]route), handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Write([]byte("info"))
	})}
}

type route struct {
	handler  http.Handler
	children map[string]route
}

func (d *route) GetHandler(route string) (http.Handler, string) {
	head, tail := shiftPath(route)
	fmt.Printf("GetHandler => %q - %q     %v\n", head, tail, d)
	if disp, ok := d.children[head]; ok {
		return disp.GetHandler(tail)
	}
	fmt.Printf("GetHandler => %v\n", d)
	return d.handler, route
}

func (d *route) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// var (
	// 	start time.Time
	// 	path  string
	// )

	// start = time.Now()
	// path = r.URL.Path

	// defer func() {
	// 	requestCount := atomic.AddUint64(&d.counter, 1) // atomic.LoadUint64(&d.counter)
	// 	color.Green("request %d: %s => %v\n", requestCount, path, time.Since(start))
	// }()

	handler, route := d.GetHandler(r.URL.Path)
	r.URL.Path = route
	handler.ServeHTTP(w, r)
}

func (r route) Register(path string, handler http.Handler) {
	fmt.Println("****register", &handler, path)
	// if path == "/" {
	// 	fmt.Println("register", &handler, r)
	// 	r.handler = handler
	// 	return
	// }

	head, tail := shiftPath(path)

	switch tail {
	case "/":
		r.children[head] = route{children: make(map[string]route), handler: handler}
	default:
		if _, ok := r.children[head]; !ok {
			r.children[head] = route{children: make(map[string]route), handler: http.NotFoundHandler()}
		}
		r.children[head].Register(tail, handler)
	}
}

// func (d *shiftPathDispatcher) registerSubRoute(head, tail string, handler http.Handler) {

// 	subDispatcher, ok := d.routes[head]
// 	if !ok {
// 		// subDispatcher =  // NewDispatcher(d.routes[head], false, d.debug)
// 		d.routes[head] = shiftPathDispatcher{routes: make(map[string]shiftPathDispatcher), roothandler: handler}
// 	}
// 	subDispatcher.Register(tail, handler)
// }

func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}
