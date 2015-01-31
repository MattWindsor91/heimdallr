package tokeniser

import "testing"

func cmpByteSlices(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i, abyte := range a {
		if abyte != b[i] {
			return false
		}
	}

	return true
}

func TestPack(t *testing.T) {
	cases := []struct {
		word string
		args []string
		want []byte
	}{
		// Unescaped command
		{
			"load",
			[]string{"/home/donald/wjaz.mp3"},
			[]byte("load /home/donald/wjaz.mp3"),
		},
		// Backslashes
		{
			"load",
			[]string{"C:\\silly\\windows\\is\\silly"},
			[]byte("load 'C:\\silly\\windows\\is\\silly'"),
		},
		// No args
		{
			"play",
			[]string{},
			[]byte("play"),
		},
		// Spaces
		{
			"load",
			[]string{"/home/donald/01 The Nightfly.mp3"},
			[]byte("load '/home/donald/01 The Nightfly.mp3'"),
		},
		// Single quotes
		{
			"foo",
			[]string{"a'bar'b"},
			[]byte("foo 'a'\\''bar'\\''b'"),
		},
		// Double quotes
		{
			"foo",
			[]string{"a\"bar\"b"},
			[]byte("foo 'a\"bar\"b'"),
		},
	}

	for _, c := range cases {
		got := Pack(c.word, c.args)
		if !cmpByteSlices(c.want, got) {
			t.Errorf("Pack(%q, %q) == %q, want %q", c.word, c.args, got, c.want)
		}
	}
}
