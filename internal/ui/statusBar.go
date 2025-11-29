package ui

import (
	"github.com/rivo/tview"
)

type statusBar struct {
	tview.Primitive
	textView *tview.TextView
}

func newStatusBar(errorChannel chan error) statusBar {
	textview := tview.NewTextView()
	textview.SetDynamicColors(true)

	go func() {
		for err := range errorChannel {
			textview.SetText("[red]" + err.Error())
		}
	}()

	return statusBar{
		Primitive: textview,
		textView:  textview,
	}
}
