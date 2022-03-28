package httpsrvr

import (
	"testing"
)

func TestShiftPath(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{input: "a/b/c", want: []string{"a", "/b/c"}},
		{input: "a/b/c////", want: []string{"a", "/b/c"}},
		{input: "abc", want: []string{"abc", "/"}},
		{input: "/a/b/c/", want: []string{"a", "/b/c"}},
		{input: "/", want: []string{"", "/"}},
		{input: "", want: []string{"", "/"}},
		{input: "//", want: []string{"", "/"}},
	}

	for _, tc := range tests {
		head, tail := ShiftPath(tc.input)
		if head != tc.want[0] || tail != tc.want[1] {
			t.Fatalf("%q: expected: %v, got: [%s %s]", tc.input, tc.want, head, tail)
		}
	}
}
