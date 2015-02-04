package baps3protocol

import "strings"

type MessageWord int

const (
	/* Message word constants.
	 * When adding to this, also add the string equivalent to LookupRequest and
	 * LookupResponse.
	 */

	BadWord MessageWord = iota

	// - Requests
	RqUnknown
	// -- Core
	RqQuit
	// -- PlayStop feature
	RqPlay
	RqStop
	// -- FileLoad feature
	RqEject
	RqLoad
	// -- Playlist feature
	RqCount
	RqDequeue
	RqEnqueue
	RqSelect

	// - Responses
	RsUnknown
	// -- Core
	RsOk
	RsFail
	RsWhat
	RsOhai
	RsFeatures
	RsState
	// -- End feature
	RsEnd
	// -- FileLoad feature
	RsFile
	// -- TimeReport feature
	RsTime
)

// Yes, a global variable.
// Go can't handle constant arrays.
var wordStrings = []string{
	"<BAD WORD>",         // BadWord
	"<UNKNOWN REQUEST>",  // RqUnknown
	"quit",               // RqQuit
	"play",               // RqPlay
	"stop",               // RqStop
	"eject",              // RqEject
	"load",               // RqLoad
	"count",              // RqCount
	"dequeue",            // RqDequeue
	"enqueue",            // RqEnqueue
	"select",             // RqSelect
	"<UNKNOWN RESPONSE>", // RsUnknown
	"OK",                 // RsOk
	"FAIL",               // RsFail
	"WHAT",               // RsWhat
	"OHAI",               // RsOhai
	"FEATURES",           // RsFeatures
	"STATE",              // RsState
	"END",                // RsEnd
	"FILE",               // RsFile
	"TIME",               // RsTime
}

func (word MessageWord) IsUnknown() bool {
	return word == BadWord || word == RqUnknown || word == RsUnknown
}

func (word MessageWord) String() string {
	return wordStrings[int(word)]
}

func LookupWord(word string) MessageWord {
	// This is O(n) on the size of WordStrings, which is unfortunate, but
	// probably ok.
	for i, str := range wordStrings {
		if str == word {
			return MessageWord(i)
		}
	}

	// In BAPS3, lowercase words are requests; uppercase words are responses.
	if strings.ToLower(word) == word {
		return RqUnknown
	} else if strings.ToUpper(word) == word {
		return RsUnknown
	}
	return BadWord
}

type message struct {
	word MessageWord
	args []string
}

func NewMessage(word MessageWord) *message {
	m := new(message)
	m.word = word
	return m
}

func (m *message) AddArg(arg string) *message {
	m.args = append(m.args, arg)
	return m
}

func (m *message) AsSlice() []string {
	slice := []string{m.word.String()}
	for _, arg := range m.args {
		slice = append(slice, arg)
	}
	return slice
}

func (m *message) Pack() []byte {
	return Pack(m.word.String(), m.args)
}
