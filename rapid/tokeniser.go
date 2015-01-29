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

func (t *Tokeniser) Push(b byte) {
	t.word.push(b)
	t.escape_next_char = false
}

func (t *Tokeniser) EndLine() {
	break
}

func (t *Tokeniser) EndWord() {
	break
}

func (t *Tokeniser) Parse(data []byte) {
	for _, b := range data {
		if t.escape_next_char {
			// TODO: Make unicode safe
			t.word.append(b)
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
				t.EndLine()
			default:
				if b == ' ' {
					t.EndWord()
				} else {
					t.Push(b)
				}
			}

		case Single:
			switch b {
			case '\'':
				t.quote_type = QuoteType.None
			default:
				t.Push(b)
			}

		case Double:
			switch b {
			case '"':
				t.quote_type = None
			case '\\':
				t.escape_next_character = true
			default:
				t.Push(b)
			}
		}
	}
}
