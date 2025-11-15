package ui

import (
	"bytes"

	"github.com/alecthomas/chroma/v2/quick"
)

func Highlight(text string, lexer string) string {
	var buf bytes.Buffer
	err := quick.Highlight(&buf, text, lexer, "terminal256", codeStyle)
	if err != nil {
		return text
	}

	return buf.String()
}
