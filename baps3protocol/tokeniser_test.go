package baps3protocol

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

	// Tests adapted from:
	// https://github.com/UniversityRadioYork/baps3-protocol.rs
	// (src/proto/unpack.rs)

	cases := []struct {
		in   string
		want [][]string
	}{
		// Empty string and empty line
		{
			"",
			[][]string{},
		},
		{
			"\n",
			[][]string{[]string{}},
		},
		// Inter-word whitespace
		{
			"foo bar baz\n",
			[][]string{[]string{"foo", "bar", "baz"}},
		},
		{
			"fizz\tbuzz\tpoo\n",
			[][]string{[]string{"fizz", "buzz", "poo"}},
		},
		// CRLF tolerance
		{
			"silly windows\r\n",
			[][]string{[]string{"silly", "windows"}},
		},
		// Leading and trailing whitespace
		{
			"     abc def\n",
			[][]string{[]string{"abc", "def"}},
		},
		{
			"ghi jkl     \n",
			[][]string{[]string{"ghi", "jkl"}},
		},
		{
			"     mno pqr     \n",
			[][]string{[]string{"mno", "pqr"}},
		},
		// Sample BAPS3 command, with double-quoted Windows path
		{
			`enqueue file "C:\\Users\\Test\\Artist - Title.mp3" 1` + "\n",
			[][]string{
				[]string{"enqueue", "file", `C:\Users\Test\Artist - Title.mp3`, "1"},
			},
		},
		// Escaped newline
		{
			"abc\\\ndef\n",
			[][]string{[]string{"abc\ndef"}},
		},
		{
			"\"abc\ndef\"\n",
			[][]string{[]string{"abc\ndef"}},
		},
		{
			"'abc\ndef'\n",
			[][]string{[]string{"abc\ndef"}},
		},
		// Multiple lines
		{
			"first line\nsecond line\n",
			[][]string{
				[]string{"first", "line"},
				[]string{"second", "line"},
			},
		},
		// Single-quote escaping
		{
			`'hello, I'\''m an escaped single quote'` + "\n",
			[][]string{[]string{"hello, I'm an escaped single quote"}},
		},
		// UTF-8
		{
			"北野 武\n",
			[][]string{[]string{"北野", "武"}},
		},
		// Not UTF-8 (ISO-8859-1).
		// Should replace bad byte with the Unicode replacement
		// character.  See example at:
		// https://en.wikipedia.org/wiki/Unicode_replacement_character
		{
			"f\xfcr\n",
			[][]string{[]string{"f\xef\xbf\xbdr"}},
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
