package util

import "testing"

func TestParsePage(t *testing.T) {
	cases := []struct {
		raw  string
		want int
	}{
		{"1", 1},
		{"0", 1},
		{"-3", 1},
		{"abc", 1},
		{"5", 5},
		{"", 1},
	}
	for _, c := range cases {
		if got := ParsePage(c.raw); got != c.want {
			t.Errorf("ParsePage(%q) = %d, want %d", c.raw, got, c.want)
		}
	}
}

func TestParsePageSize(t *testing.T) {
	cases := []struct {
		raw  string
		want int
	}{
		{"10", 10},
		{"0", 10},
		{"300", 200},
		{"abc", 10},
	}
	for _, c := range cases {
		if got := ParsePageSize(c.raw); got != c.want {
			t.Errorf("ParsePageSize(%q) = %d, want %d", c.raw, got, c.want)
		}
	}
}
