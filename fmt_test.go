package tblfmt

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	runewidth "github.com/mattn/go-runewidth"
)

func TestTabwidth(t *testing.T) {
	tests := []struct {
		s      string
		offset int
		tab    int
		exp    int
	}{
		{"", 0, 8, 0},
		{" ", 0, 8, 1},
		{"    ", 0, 8, 4},
		{"\u8888", 0, 8, 2},
		{"\t", 0, 8, 8},
		{"\t\t", 0, 8, 16}, // 5
		{"\t\t ", 0, 8, 17},
		{" \t\t ", 0, 8, 17},
		{" \t\t\t ", 0, 8, 25},
		{"foo\tbar\t", 0, 8, 16},
		{"\t\t\u8888", 0, 8, 18}, // 10
		{"\u8888\t\u8888", 0, 8, 10},
		{"\u8888\t\u8888\t", 0, 8, 16},

		{"", 1, 8, 0}, // 13
		{"\t", 1, 4, 3},
		{" \t", 1, 4, 3},
		{" \t ", 1, 4, 4},
		{"\u8888\t\u8888\t", 1, 4, 7},

		{"\u8888\t\t\u8888\t", 3, 2, 9}, // 18
		/*
		   ---xxxxxxxxx (width == 9)
		  |   袈   袈  |
		*/

		{"\u8888\t\u8888\t\t\u8888", 14, 8, 28}, // 19
		/*
		   --------------xxxxxxxxxxxxxxxxxxxxxxxxxxxx (width == 28)
		  |              袈        袈              袈
		*/
		{"袈	袈		袈", 14, 8, 28}, // 19
	}

	for i, test := range tests {
		tabs, w := tabpositions(test.s)
		w += tabwidth(tabs, test.offset, test.tab)
		if test.exp != w {
			t.Errorf(
				"test %d %q expected tabwidth(%v, %d, %d) = %d, got: %d",
				i, test.s, tabs, test.offset, test.tab, test.exp, w,
			)
		}
	}
}

var tabRE = regexp.MustCompile("\t")

// tabpositions returns a list of tab positions in s.
func tabpositions(s string) ([][2]int, int) {
	var tabs [][2]int
	var last int
	for _, m := range tabRE.FindAllStringIndex(s, -1) {
		tabs = append(tabs, [2]int{m[0], runewidth.StringWidth(s[last:m[0]])})
		last = m[0] + 1
	}
	return tabs, runewidth.StringWidth(s[last:])
}

func TestEscape(t *testing.T) {
	tests := []escTest{
		V("", 0),
		V("\u8888\t\u8888", 4),
		V(" \u8888 \t \u8888 ", 8),
	}
	for i, test := range tests {
		v := escape([]byte(test.s), nil, 0)
		if !reflect.DeepEqual(v, test.exp) {
			t.Errorf(
				"test %d %q expected %v, got: %v",
				i, test.s, test.exp, v,
			)

			width := runewidth.StringWidth(string(v.Buf))
			if v.Width != width {
				t.Errorf(
					"test %d %q expected width %d, got: %d",
					i, test.s, width, v.Width,
				)
			}

			if width != test.check {
				t.Errorf(
					"test %d %q expected check width %d, got: %d",
					i, test.s, test.check, width,
				)
			}
		}
	}
}

type escTest struct {
	s     string
	exp   *Value
	check int
}

func quote(s string) string {
	s = strconv.Quote(s)
	s = s[1 : len(s)-1]
	s = strings.Replace(s, `\t`, "\t", -1)
	s = strings.Replace(s, `\n`, "\n", -1)
	return s
}

func V(s string, check int) escTest {
	c := quote(s)
	tabs, width := tabpositions(c)
	var buf []byte
	if len(c) != 0 {
		buf = []byte(c)
	}
	v := &Value{
		Buf:   buf,
		Tabs:  [][][2]int{tabs},
		Width: width,
	}
	return escTest{s, v, check}
}
