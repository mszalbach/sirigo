// Package ui contains all elements for the TUI
package ui

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/rivo/tview"
)

// queueUpdateDrawer allows a component to use tview async safe updating of components.
// Requiered for async routines and to allow mocking this in tests since a tview.Application is not so gut testable (at my current understanding)
type queueUpdateDrawer interface {
	QueueUpdateDraw(f func()) *tview.Application
}

// NewSiriApp creates the tview application to interact with a SIRI server
func NewSiriApp(
	siriClient siri.Client,
	sendTemplates siri.TemplateCache,
	responseTemplates siri.TemplateCache,
	cancel context.CancelFunc,
) *tview.Application {
	app := tview.NewApplication()
	app.SetTitle(fmt.Sprintf("Sirigo (%s)", siriClient.ClientRef))

	initStyles()
	app.EnableMouse(true)
	app.EnablePaste(true)

	// Building UI elements
	errorChannel := make(chan error, 5)
	statusBar := newStatusBar(app, errorChannel)
	keymap := newKeymap()
	siriClientView := newSiriClientView(siriClient, sendTemplates, errorChannel)
	siriServerView := newSiriServerView(app, siriClient, responseTemplates, errorChannel)

	// Building layout
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

	// Installing shortcuts
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlX:
			cancel()
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
