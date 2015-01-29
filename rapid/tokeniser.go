package rapid

type QuoteType int

const (
	None QuoteType = iota
	Single
	Double
)

type Tokeniser struct {
	escape_next_char bool
	quote_type       QuoteType
	word             []byte
	words            []string
	lines            [][]string
}

func (t *Tokeniser) push(b byte) {
	t.word = append(t.word, b)
	t.escape_next_char = false
}

func (t *Tokeniser) endLine() {
}

func (t *Tokeniser) endWord() {
}

func (t *Tokeniser) Parse(data []byte) [][]string {
	for _, b := range data {
		if t.escape_next_char {
			// TODO: Make unicode safe
			t.push(b)
			t.escape_next_char = true
		}
		switch t.quote_type {
		case None:
			switch b {
			case '\'':
				t.quote_type = None
			case '"':
				t.quote_type = None
			case '\\':
				t.escape_next_char = true
			case '\n':
				t.endLine()
			default:
				if b == ' ' {
					t.endWord()
				} else {
					t.push(b)
				}
			}

		case Single:
			switch b {
			case '\'':
				t.quote_type = None
			default:
				t.push(b)
			}

		case Double:
			switch b {
			case '"':
				t.quote_type = None
			case '\\':
				t.escape_next_char = true
			default:
				t.push(b)
			}
		}
	}

	lines := t.lines
	t.lines = [][]string{}
	return lines
}
