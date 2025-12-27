package ui

import "github.com/rivo/tview"

type codeTextView struct {
	*tview.TextView
	app tuiApp
}

func newCodeTextView(app tuiApp, title string) *codeTextView {
	codeTextView := &codeTextView{tview.NewTextView(), app}
	codeTextView.SetDynamicColors(true).SetBorder(true).SetTitle(title)
	return codeTextView
}

func (ctv *codeTextView) SetCode(code string, language string) {
	ctv.ScrollToBeginning()
	ctv.SetText(code)
	// highlight takes a lot of time for big responses, so doing it delayed later
	go func() {
		highlighted := tview.TranslateANSI(highlight(code, language))
		ctv.app.QueueUpdateDraw(func() {
			ctv.SetText(highlighted)
		})
	}()
}
