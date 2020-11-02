package kunst

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

func TestHello(t *testing.T) {
	want := "/eins/zwei/drei/vier/fünf.sechs"
	fmt.Println(want)
	for i := len(want); want != ""; i = strings.LastIndex(want, "/") {
		want = want[:i]
		fmt.Println(i, want)
	}

}

func TestHello2(t *testing.T) {
	acl := map[string]string{"/eins/zwei": "R", "/public": "R", "/ihleven": "R"}
	want := "/eins/zwei/eins/zwei/drei/vier/fünf.sechs"

	for i := want; len(i) > 1; i = filepath.Dir(i) {

		fmt.Println("path:", i, acl[i])
	}

}
