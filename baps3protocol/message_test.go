package baps3protocol

import "testing"

func TestMessageWord(t *testing.T) {
	cases := []struct {
		str     string
		word    MessageWord
		unknown bool
	}{
		// Ok, a request
		{"load", RqLoad, false},
		// Ok, a response
		{"OHAI", RsOhai, false},
		// Unknown, but a request
		{"uwot", RqUnknown, true},
		// Unknown, but a response
		{"MATE", RsUnknown, true},
		// Unknown, and unclear what type of message
		{"MaTe", BadWord, true},
	}

	for _, c := range cases {
		gotword := LookupWord(c.str)
		if gotword != c.word {
			t.Errorf("LookupWord(%q) == %q, want %q", c.str, gotword, c.word)
		}
		if c.word.IsUnknown() != c.unknown {
			t.Errorf("%q.IsUnknown() == %q, want %q", c.word, !c.unknown, c.unknown)
		}

		// Only do the other direction if it's a valid response
		if !c.unknown {
			gotstr := c.word.String()
			if gotstr != c.str {
				t.Errorf("%q.String() == %q, want %q", c.word, gotstr, c.str)
			}
		}
	}
}
