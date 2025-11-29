// Package ui contains all elements for the TUI
package ui

import (
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/rivo/tview"
)

// NewSiriApp creates the tview app to interact with a SIRI server
func NewSiriApp(
	siriClient siri.Client,
	sendTemplates siri.TemplateCache,
	responseTemplates siri.TemplateCache,
) *tview.Application {
	app := tview.NewApplication()

	initStyles()
	app.EnableMouse(true)
	app.EnablePaste(true)

	// building ui elements
	errorChannel := make(chan error, 5)
	statusBar := newStatusBar(errorChannel)
	keymap := newKeymap()
	siriClientView := newSiriClientView(siriClient, sendTemplates, errorChannel)
	siriServerView := newSiriServerView(siriClient, responseTemplates, errorChannel)

	// building layout
	bodyFlex := tview.NewFlex().
		AddItem(siriClientView, 0, 1, false).
		AddItem(siriServerView, 0, 1, false)

	footerFlex := tview.NewFlex().
		AddItem(keymap, 0, 1, false).AddItem(statusBar, 0, 1, false)

	appFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(bodyFlex, 0, 1, false).
		AddItem(footerFlex, 2, 0, false)
	app.SetRoot(appFlex, true)

	// installing shortcuts
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlX:
			app.Stop()
			// not the correct way to shutdown app and webserver, but works for now
			os.Exit(0)
		case tcell.KeyCtrlO:
			response := siriClientView.send()
			siriServerView.setResponse(response)
			return nil
		case tcell.KeyCtrlC:
			return nil
		}
		return event
	})

	return app
}
