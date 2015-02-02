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
	RsOhai
	RsFeatures
	RsState
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
	"OHAI",               // RsOhai
	"FEATURES",           // RsFeatures
	"STATE",              // RsState
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
