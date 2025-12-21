package ui

import (
	"github.com/rivo/tview"
)

type statusBar struct {
	*tview.TextView
}

func newStatusBar(app tuiApp, errorChannel chan error) statusBar {
	textview := tview.NewTextView()
	textview.SetDynamicColors(true)

	go func() {
		for err := range errorChannel {
			app.QueueUpdateDraw(func() {
				textview.SetText("[red]" + err.Error())
			})
		}
	}()

	return statusBar{
		textview,
	}
}
