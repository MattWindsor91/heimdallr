package baps3protocol

import "testing"

// cmpWords is defined in tokeniser_test.
// TODO(CaptainHayashi): move cmpWords elsewhere?

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

func TestMessage(t *testing.T) {
	cases := []struct {
		words []string
		msg   *Message
	}{
		// Empty request
		{[]string{"play"}, NewMessage(RqPlay)},
		// Request with one argument
		{[]string{"load", "foo"}, NewMessage(RqLoad).AddArg("foo")},
		// Request with multiple argument
		{[]string{"enqueue", "0", "file", "blah"},
			NewMessage(RqEnqueue).AddArg("0").AddArg("file").AddArg("blah"),
		},
		// Empty response
		{[]string{"END"}, NewMessage(RsEnd)},
		// Response with one argument
		{[]string{"FILE", "foo"}, NewMessage(RsFile).AddArg("foo")},
		// Response with multiple argument
		{[]string{"FAIL", "nou", "load", "blah"},
			NewMessage(RsFail).AddArg("nou").AddArg("load").AddArg("blah"),
		},
	}

	for _, c := range cases {
		gotslice := c.msg.AsSlice()
		if !cmpWords(gotslice, c.words) {
			t.Errorf("%q.ToSlice() == %q, want %q", c.msg, gotslice, c.words)
		}
	}
}
