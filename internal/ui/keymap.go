package ui

import (
	"strings"

	"github.com/rivo/tview"
)

var keys = []struct {
	key         string
	description string
}{
	{
		key:         "F1",
		description: "Help",
	},
	{
		key:         "Ctrl-O",
		description: "Send",
	},
	{
		key:         "Ctrl-E",
		description: "Editor",
	},
	{
		key:         "Ctrl-X",
		description: "Exit",
	},
}

func newKeymap() *tview.TextView {
	keyMap := tview.NewTextView()
	keyMap.SetDynamicColors(true)

	builder := strings.Builder{}
	for _, k := range keys {
		builder.WriteString(keyColor)
		builder.WriteString(k.key)
		builder.WriteString(descriptionColor)
		builder.WriteString(" ")
		builder.WriteString(k.description)
		builder.WriteString(" ")
	}
	keyMap.SetText(builder.String())
	return keyMap
}
