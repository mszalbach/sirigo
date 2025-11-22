package ui

import (
	"bytes"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// copied from k9s dracula style
var colors = map[string]tcell.Color{
	"foreground": tcell.GetColor("#f8f8f2"),
	"background": tcell.GetColor("#282a36"),
	"selection":  tcell.GetColor("#44475a"),
	"purple":     tcell.GetColor("#bd93f9"),
	"orange":     tcell.GetColor("#ffb86c"),
	"yellow":     tcell.GetColor("#f1fa8c"),
	"pink":       tcell.GetColor("#ff79c6"),
	"comment":    tcell.GetColor("#6272a4"),
}

const codeStyle = "dracula"

func InitStyles() {
	tview.Styles.PrimaryTextColor = colors["foreground"]
	tview.Styles.SecondaryTextColor = colors["orange"]
	tview.Styles.TitleColor = colors["purple"]
	tview.Styles.BorderColor = colors["selection"]
	tview.Styles.PrimitiveBackgroundColor = colors["background"]
	tview.Styles.ContrastBackgroundColor = colors["selection"]
	tview.Styles.ContrastSecondaryTextColor = colors["comment"]

	// does not change anything currently used
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorGreen
	tview.Styles.GraphicsColor = tcell.ColorGreen
	tview.Styles.TertiaryTextColor = tcell.ColorGreen
	tview.Styles.InverseTextColor = tcell.ColorGreen
}

func Highlight(text string, lexer string) string {
	var buf bytes.Buffer
	err := quick.Highlight(&buf, text, lexer, "terminal256", codeStyle)
	if err != nil {
		return text
	}

	return buf.String()
}
