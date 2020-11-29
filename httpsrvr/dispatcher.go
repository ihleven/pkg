package httpsrvr

import (
	"net/http"
	"path"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
)

func NewDispatcher(notFoundHandler http.Handler, isRoot bool, debug bool) *shiftPathDispatcher {

	routes := make(map[string]http.Handler)

	return &shiftPathDispatcher{routes, notFoundHandler, isRoot, debug, 0}
}

type shiftPathDispatcher struct {
	routes          map[string]http.Handler
	notFoundHandler http.Handler
	isRoot          bool
	debug           bool
	counter         uint64
}

func (d *shiftPathDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var (
		start            time.Time
		path, head, tail string
	)

	if d.isRoot {
		start = time.Now()
		atomic.AddUint64(&d.counter, 1)
		path = r.URL.Path

		// log.NewRequest(reqNumber, start, r)

		// values := r.URL.Query()
		// values.Add("_reqNumber", strconv.FormatInt(reqNumber, 10))
		// r.URL.RawQuery = values.Encode()
	}
	// fmt.Println("dispatch => ", r.URL.Path)
	head, tail = shiftPath(r.URL.Path)
	// if d.debug {
	// 	color.Green(" * dispatching: %s -> %s\n", head, tail)
	// }
	route, ok := d.routes[head]
	switch {
	case ok:
		r.URL.Path = tail
		route.ServeHTTP(w, r)

	case d.notFoundHandler != nil:
		d.notFoundHandler.ServeHTTP(w, r)
	default:
		http.NotFound(w, r)
	}

	if d.isRoot {
		requestCount := atomic.LoadUint64(&d.counter)
		if d.debug {
			color.Green(" * %s   => %v,  request count: %d\n", path, time.Since(start), requestCount)
		}
	}
}

func (d *shiftPathDispatcher) Register(route string, handler http.Handler) {

	var head, tail string
	head, tail = shiftPath(route)

	switch tail {
	case "/":
		d.routes[head] = handler
	default:
		d.registerSubRoute(head, tail[1:], handler)
	}
}

func (d *shiftPathDispatcher) registerSubRoute(head, tail string, handler http.Handler) {

	subDispatcher, ok := d.routes[head].(*shiftPathDispatcher)
	if !ok {
		subDispatcher = NewDispatcher(d.routes[head], false, d.debug)
		d.routes[head] = subDispatcher
	}
	subDispatcher.Register(tail, handler)
}

func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}
