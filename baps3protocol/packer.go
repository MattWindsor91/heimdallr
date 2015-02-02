package baps3protocol

import (
	"bytes"
	"strings"
	"unicode"
)

// Pack a command word and 0+ args into a slice of bytes ready for sending.
// Args will be single quote escaped if they contain ' " \ or whitespace
func Pack(word string, args []string) []byte {
	output := new(bytes.Buffer)
	output.WriteString(word)
	for _, a := range args {
		// Escape arg if needed
		for _, c := range a {
			if c < unicode.MaxASCII && (unicode.IsSpace(c) || strings.ContainsRune(`'"\`, c)) {
				a = escapeArgument(a)
				break
			}
		}
		output.WriteString(" " + a)
	}
	return output.Bytes()
}

func escapeArgument(input string) (output string) {
	return "'" + strings.Replace(input, "'", `'\''`, -1) + "'"
}
