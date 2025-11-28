package ui

import (
	"strings"

	"github.com/rivo/tview"
)

var (
	descriptionColor = colorTag(colors["foreground"], colors["background"])
	keyColor         = colorTag(colors["background"], colors["selection"])

	keys = []struct {
		key         string
		description string
	}{
		{
			key:         "^O",
			description: "Send",
		},
		{
			key:         "^X",
			description: "Exit",
		},
	}
)

func newKeymap() *tview.TextView {
	keyMap := tview.NewTextView()
	keyMap.SetDynamicColors(true)

	builder := strings.Builder{}
	for _, k := range keys {
		builder.WriteString(keyColor + k.key + descriptionColor + " " + k.description + " ")
	}
	keyMap.SetText(builder.String())
	return keyMap
}
