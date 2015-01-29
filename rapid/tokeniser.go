package rapid

import "bytes"

type QuoteType int

const (
	None QuoteType = iota
	Single
	Double
)

type Tokeniser struct {
	escape_next_char bool
	quote_type       QuoteType
	word             *bytes.Buffer
	words            []string
	lines            [][]string
}

func (t *Tokeniser) endLine() {
	// We might still be in the middle of a word
	t.endWord()

	t.lines = append(t.lines, t.words)
	t.words = []string{}
}

func (t *Tokeniser) endWord() {
	t.words = append(t.words, t.word.String())
	t.word.Truncate(0)
}

func (t *Tokeniser) Parse(data []byte) [][]string {
	for _, b := range data {
		if t.escape_next_char {
			// TODO: Make unicode safe
			t.word.WriteByte(b)
			t.escape_next_char = false
		}
		switch t.quote_type {
		case None:
			switch b {
			case '\'':
				t.quote_type = Single
			case '"':
				t.quote_type = Double
			case '\\':
				t.escape_next_char = true
			case '\n':
				t.endLine()
			default:
				if b == ' ' {
					t.endWord()
				} else {
					t.word.WriteByte(b)
				}
			}

		case Single:
			switch b {
			case '\'':
				t.quote_type = None
			default:
				t.word.WriteByte(b)
			}

		case Double:
			switch b {
			case '"':
				t.quote_type = None
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
