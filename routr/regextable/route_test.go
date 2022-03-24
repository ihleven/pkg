package regextable

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNewRouter(t *testing.T) {

	contact := func(w http.ResponseWriter, r *http.Request, p map[string]string) {
		fmt.Fprintf(w, "contact %v\n", p)
	}
	router := NewRouter()
	router.Register("GET", "/bilder/{id:int}", contact)
	router.Register("GET", "/slugs/{slug}/{id:int}", contact)
	router.Register("GET", "/hello/(?P<first>[^/]+)", contact)
	// http.Handler("/", Serve)
	// http.ListenAndServe(":9000", router)
}
