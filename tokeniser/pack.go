package baps3protocol

import (
	"bytes"
	"strings"
)

func Pack(word string, args []string) []byte {
	output := new(bytes.Buffer)
	// TODO: check command word is valid
	output.WriteString(word)
	for _, a := range args {
		if strings.ContainsAny(a, "'\"\\ ") { // Does this arg contain special characters?
			a = escapeArgument(a)
		}
		output.WriteString(" " + a)
	}
	return output.Bytes()
}

func escapeArgument(input string) (output string) {
	return "'" + strings.Replace(input, "'", "'\\''", -1) + "'"
}
