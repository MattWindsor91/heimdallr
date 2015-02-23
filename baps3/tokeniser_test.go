package baps3

import (
	"testing"
)

func cmpLines(a [][]string, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, aline := range a {
		if !cmpWords(aline, b[i]) {
			return false
		}
	}

	return true
}

func cmpWords(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, aword := range a {
		if aword != b[i] {
			return false
		}
	}

	return true

}

func TestTokenise(t *testing.T) {
	// For now, only test one complete line at a time.
	// TODO(CaptainHayashi): add partial-line tests.

	// Tests adapted from (and labelled with respect to):
	// http://universityradioyork.github.io/baps3-spec/comms/internal/protocol.html#examples

	cases := []struct {
		in   string
		want [][]string
	}{
		// E1 - empty string
		{
			"",
			[][]string{},
		},
		// E2 - empty line
		{
			"\n",
			[][]string{[]string{}},
		},
		// E3 - empty single-quoted string
		{
			"''\n",
			[][]string{[]string{""}},
		},
		// E4 - empty double-quoted string
		{
			"\"\"\n",
			[][]string{[]string{""}},
		},
		// W1 - space-delimited words
		{
			"foo bar baz\n",
			[][]string{[]string{"foo", "bar", "baz"}},
		},
		// W2 - tab-delimited words
		{
			"fizz\tbuzz\tpoo\n",
			[][]string{[]string{"fizz", "buzz", "poo"}},
		},
		// W3 - oddly-delimited words
		{
			"bibbity\rbobbity\rboo\n",
			[][]string{[]string{"bibbity", "bobbity", "boo"}},
		},
		// W4 - CRLF tolerance
		{
			"silly windows\r\n",
			[][]string{[]string{"silly", "windows"}},
		},
		// W5 - leading whitespace
		{
			"     abc def\n",
			[][]string{[]string{"abc", "def"}},
		},
		// W6 - trailing whitespace
		{
			"ghi jkl     \n",
			[][]string{[]string{"ghi", "jkl"}},
		},
		// W7 - surrounding whitespace
		{
			"     mno pqr     \n",
			[][]string{[]string{"mno", "pqr"}},
		},
		// Q1 - backslash escaping
		{
			"abc\\\ndef\n",
			[][]string{[]string{"abc\ndef"}},
		},
		// Q2 - double-quoting
		{
			"\"abc\ndef\"\n",
			[][]string{[]string{"abc\ndef"}},
		},
		// Q3 - double-quoting, backslash-escape
		{
			"\"abc\\\ndef\"\n",
			[][]string{[]string{"abc\ndef"}},
		},
		// Q4 - single-quoting
		{
			"'abc\ndef'\n",
			[][]string{[]string{"abc\ndef"}},
		},
		// Q5 - single-quoting, backslash-'escape'
		{
			"'abc\\\ndef'\n",
			[][]string{[]string{"abc\\\ndef"}},
		},
		// Q6 - backslash-escaped double quote
		{
			"Scare\\\" quotes\\\"\n",
			[][]string{[]string{"Scare\"", "quotes\""}},
		},
		// Q7 - backslash-escaped single quote
		{
			"I\\'m free\n",
			[][]string{[]string{"I'm", "free"}},
		},
		// Q8 - single-quoted single quote
		{
			`'hello, I'\''m an escaped single quote'` + "\n",
			[][]string{[]string{"hello, I'm an escaped single quote"}},
		},
		// Q9 - double-quoted single quote
		{
			`"hello, this is an \" escaped double quote"` + "\n",
			[][]string{[]string{`hello, this is an " escaped double quote`}},
		},
		// M1 - multiple lines
		{
			"first line\nsecond line\n",
			[][]string{
				[]string{"first", "line"},
				[]string{"second", "line"},
			},
		},
		// U1 - UTF-8
		{
			"北野 武\n",
			[][]string{[]string{"北野", "武"}},
		},
		// U2 - Not UTF-8 (ISO-8859-1).
		// Should replace bad byte with the Unicode replacement
		// character.  See example at:
		// https://en.wikipedia.org/wiki/Unicode_replacement_character
		{
			"f\xfcr\n",
			[][]string{[]string{"f\xef\xbf\xbdr"}},
		},
		// X1 - Sample BAPS3 command, with double-quoted Windows path
		{
			`enqueue file "C:\\Users\\Test\\Artist - Title.mp3" 1` + "\n",
			[][]string{
				[]string{"enqueue", "file", `C:\Users\Test\Artist - Title.mp3`, "1"},
			},
		},
	}

	for _, c := range cases {
		tok := NewTokeniser()
		got, _, err := tok.Tokenise([]byte(c.in))
		if err != nil {
			t.Errorf("Tokenise(%q) gave error %q", c.in, err)
		}
		if !cmpLines(got, c.want) {
			t.Errorf("Tokenise(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
