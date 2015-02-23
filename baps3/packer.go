package baps3

import (
	"bytes"
	"strings"
	"unicode"
)

// Pack a command word and 0+ args into a slice of bytes ready for sending.
// Args will be single quote escaped if they contain ' " \ or whitespace
// Pack may return an error if the resulting byte-slice is too large.
func Pack(word string, args []string) (packed []byte, err error) {
	packed, err = []byte{}, nil

	output := new(bytes.Buffer)

	_, err = output.WriteString(word)
	if err != nil {
		return
	}

	for _, a := range args {
		// Escape arg if needed
		for _, c := range a {
			if c < unicode.MaxASCII && (unicode.IsSpace(c) || strings.ContainsRune(`'"\`, c)) {
				a = escapeArgument(a)
				break
			}
		}

		_, err = output.WriteString(" " + a)
		if err != nil {
			return
		}
	}

	packed = output.Bytes()
	return
}

func escapeArgument(input string) (output string) {
	return "'" + strings.Replace(input, "'", `'\''`, -1) + "'"
}
