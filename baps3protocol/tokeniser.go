package baps3protocol

import (
	"bytes"
	"unicode"
)

// quoteType represents one of the types of quoting used in the BAPS3 protocol.
type quoteType int

const (
	// none represents the state between quoted parts of a BAPS3 message.
	none quoteType = iota

	// single represents 'single quoted' parts of a BAPS3 message.
	single

	// double represents "double quoted" parts of a BAPS3 message.
	double
)

// Tokeniser holds the state of a BAPS3 protocol tokeniser.
type Tokeniser struct {
	escape_next_char bool
	quote_type       quoteType
	word             *bytes.Buffer
	words            []string
	lines            [][]string
}

// NewTokeniser creates and returns a new, empty Tokeniser.
func NewTokeniser() *Tokeniser {
	t := new(Tokeniser)
	t.escape_next_char = false
	t.quote_type = none
	t.word = new(bytes.Buffer)
	t.words = []string{}
	t.lines = [][]string{}
	return t
}

func (t *Tokeniser) endLine() {
	// We might still be in the middle of a word.
	t.endWord()

	t.lines = append(t.lines, t.words)
	t.words = []string{}
}

func (t *Tokeniser) endWord() {
	if t.word.Len() == 0 {
		// Don't add an empty word.
		return
	}

	// This ensures any non-UTF8 is replaced with the Unicode replacement
	// character.  We could use String(), but this would permit invalid
	// UTF8.
	uword := []rune{}
	for {
		r, _, err := t.word.ReadRune()
		if err != nil {
			break
		}
		uword = append(uword, r)
	}

	t.words = append(t.words, string(uword))
	t.word.Truncate(0)
}

// Tokenise feeds raw bytes into a Tokeniser.
// If the bytes include the ending of one or more command lines, those lines
// shall be returned, as a slice of lines represented as slices of
// word-strings.  Else, the slice shall be empty.
func (t *Tokeniser) Tokenise(data []byte) [][]string {
	for _, b := range data {
		if t.escape_next_char {
			t.word.WriteByte(b)
			t.escape_next_char = false
			continue
		}

		switch t.quote_type {
		case none:
			switch b {
			case '\'':
				t.quote_type = single
			case '"':
				t.quote_type = double
			case '\\':
				t.escape_next_char = true
			case '\n':
				t.endLine()
			default:
				// Note that this will only check for ASCII
				// whitespace, because we only pass it one byte
				// and non-ASCII whitespace is >1 UTF-8 byte.
				if unicode.IsSpace(rune(b)) {
					t.endWord()
				} else {
					t.word.WriteByte(b)
				}
			}

		case single:
			switch b {
			case '\'':
				t.quote_type = none
			default:
				t.word.WriteByte(b)
			}

		case double:
			switch b {
			case '"':
				t.quote_type = none
			case '\\':
				t.escape_next_char = true
			default:
				t.word.WriteByte(b)
			}
		}
	}

	lines := t.lines
	t.lines = [][]string{}
	return lines
}
