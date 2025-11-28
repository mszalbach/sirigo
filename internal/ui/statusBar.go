package ui

import (
	"github.com/rivo/tview"
)

type statusBar struct {
	tview.Primitive
	textView *tview.TextView
}

func newStatusBar() statusBar {
	textview := tview.NewTextView()
	textview.SetDynamicColors(true)

	return statusBar{
		Primitive: textview,
		textView:  textview,
	}
}

func (f statusBar) error(errorMessage string) {
	f.textView.SetText("[red]" + errorMessage)
}
