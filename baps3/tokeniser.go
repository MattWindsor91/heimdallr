package baps3

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
	inWord           bool
	escapeNextChar   bool
	currentQuoteType quoteType
	word             *bytes.Buffer
	words            []string
	lines            [][]string
	err              error
}

// NewTokeniser creates and returns a new, empty Tokeniser.
func NewTokeniser() *Tokeniser {
	t := new(Tokeniser)
	t.escapeNextChar = false
	t.currentQuoteType = none
	t.word = new(bytes.Buffer)
	t.inWord = false
	t.words = []string{}
	t.lines = [][]string{}
	t.err = nil
	return t
}

func (t *Tokeniser) endLine() {
	// We might still be in the middle of a word.
	t.endWord()

	t.lines = append(t.lines, t.words)
	t.words = []string{}
}

func (t *Tokeniser) endWord() {
	if !t.inWord {
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
	t.inWord = false
}

// Tokenise feeds raw bytes into a Tokeniser.
//
// If the bytes include the ending of one or more command lines, those lines
// shall be returned, as a slice of lines represented as slices of
// word-strings.  Else, the slice shall be empty.
//
// Tokenise may return an error if its current word gets over-full.  In this
// case, lines will contain the lines it managed to tokenise before keeling
// over, count will contain the number of bytes processed (including the byte
// causing the error), and the tokeniser will remain in error until replaced or
// the current word is ended.
func (t *Tokeniser) Tokenise(data []byte) (lines [][]string, count uint64, err error) {
	count = 0
	for _, b := range data {
		// Exit early if the current word has become over-full.  This
		// lets the caller handle the error without silently masking it
		// if we manage to progress.
		if t.err != nil {
			break
		}

		if t.escapeNextChar {
			t.put(b)
			t.escapeNextChar = false
			continue
		}

		switch t.currentQuoteType {
		case none:
			t.tokeniseNoQuotes(b)
		case single:
			t.tokeniseSingleQuotes(b)
		case double:
			t.tokeniseDoubleQuotes(b)
		}

		count++
	}

	lines, t.lines = t.lines, [][]string{}
	err, t.err = t.err, nil

	return
}

// tokeniseNoQuotes tokenises a single byte outside quote characters.
func (t *Tokeniser) tokeniseNoQuotes(b byte) {
	switch b {
	case '\'':
		// Switching into single quotes mode starts a word.
		// This is to allow '' to represent the empty string.
		t.inWord = true
		t.currentQuoteType = single
	case '"':
		// Switching into double quotes mode starts a word.
		// This is to allow "" to represent the empty string.
		t.inWord = true
		t.currentQuoteType = double
	case '\\':
		t.escapeNextChar = true
	case '\n':
		t.endLine()
	default:
		// Note that this will only check for ASCII
		// whitespace, because we only pass it one byte
		// and non-ASCII whitespace is >1 UTF-8 byte.
		if unicode.IsSpace(rune(b)) {
			t.endWord()
		} else {
			t.put(b)
		}
	}
}

// tokeniseSingleQuotes tokenises a single byte within single quotes.
func (t *Tokeniser) tokeniseSingleQuotes(b byte) {
	switch b {
	case '\'':
		t.currentQuoteType = none
	default:
		t.put(b)
	}
}

// tokeniseDoubleQuotes tokenises a single byte within double quotes.
func (t *Tokeniser) tokeniseDoubleQuotes(b byte) {
	switch b {
	case '"':
		t.currentQuoteType = none
	case '\\':
		t.escapeNextChar = true
	default:
		t.put(b)
	}
}

// put adds a byte to the Tokeniser's buffer.
// If the buffer is too big, an error will be raised and propagated to the
// Tokeniser's user.
func (t *Tokeniser) put(b byte) {
	t.err = t.word.WriteByte(b)
	t.inWord = true
}
