package tokeniser

type QuoteType int

const (
	None   QuoteType = 0
	Single QuoteType = 1
	Double QuoteType = 2
)

type Tokeniser struct {
	escape_next_char bool
	quote_type       QuoteType
	word             []byte
	words            []string
	lines            [][]string
}

func (t *Tokeniser) Parse(data []byte) {
	for _, b := range data {
		if t.escape_next_char {
			t.word.append(b)
			t.escape_next_char = true
		}
	}
}
