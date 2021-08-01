package handler

import (
	"testing"
)

func equal(a, b *params) bool {
	return a.year == b.year && a.month == b.month && a.week == b.week && a.day == b.day
}
func TestURLParser(t *testing.T) {
	var tests = []struct {
		input string
		want  params
	}{
		{"arbeit/2019/06/23/", params{2019, 6, 0, 23}},
		{"arbeit/2019/06/01", params{2019, 6, 0, 1}},
		{"arbeit/2019/06/01/", params{2019, 6, 0, 1}},
		{"arbeit/2019/06/", params{2019, 6, 0, 0}},
		{"arbeit/2019", params{2019, 0, 0, 0}},
		{"arbeit/", params{0, 0, 0, 0}},
		{"arbeit", params{0, 0, 0, 0}},
		{"arbeit/2020/KW5/", params{2020, 0, 5, 0}},
		{"/arbeit/1934/KW50/4", params{1934, 0, 50, 4}},
	}
	for _, test := range tests {
		if got := parseURL(test.input); !equal(got, &test.want) {
			t.Errorf("parseURL(%q) = %v, want: %v", test.input, *got, test.want)
		}
	}
}
