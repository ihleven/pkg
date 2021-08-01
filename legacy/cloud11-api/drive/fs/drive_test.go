package fs

import (
	"testing"

	"github.com/ihleven/cloud11-api/drive"
)

func equal(a drive.Handle, b *handle) bool {
	c := a.(*handle)
	return c.location == b.location && c.mode == b.mode && a.Name() == b.Name()
}
func TestGetHandle(t *testing.T) {

	var drive = FSWebDrive{Root: "/Users/mi/tmp", Prefix: "/home", ServeURL: "/serve/home", PermissionMode: 0644}

	var tests = []struct {
		input string
		want  *handle
	}{
		{"/home/14/DSC02007.txt", &handle{location: "/Users/mi/tmp/14/DSC02007.txt", mode: 0644}},
	}
	for _, test := range tests {
		t.Log(drive.GetHandle(test.input))
		if got, err := drive.GetHandle(test.input); !equal(got, test.want) {
			t.Log(got, err)
			t.Errorf("parseURL(%q) = %v, want: %v", test.input, got.(*handle), test.want)
		}
	}
}
