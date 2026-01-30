// Package ui contains all elements for the TUI
package ui

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/rivo/tview"
)

// tuiApp allows a component to use features of the tview.Application and custom provided ones.
// Requiered for mocking this in tests since a tview.Application is not so good testable (at my current understanding)
type tuiApp interface {
	QueueUpdateDraw(f func()) *tview.Application
	register(prioritizedComponents ...tview.Primitive)
	Suspend(func()) bool
}

// SiriApp is the main tview application for the SIRI client
type SiriApp struct {
	*tview.Application
	focusComponents []tview.Primitive
}

// NewSiriApp creates the tview application to interact with a SIRI server
func NewSiriApp(
	siriClient *siri.Client,
	sendTemplates siri.TemplateCache,
	responseTemplates siri.TemplateCache,
	cancel context.CancelCauseFunc,
) *SiriApp {
	siriApp := &SiriApp{
		Application:     tview.NewApplication(),
		focusComponents: []tview.Primitive{},
	}
	siriApp.SetTitle(fmt.Sprintf("Sirigo (%s)", siriClient.ClientRef))

	initStyles()
	siriApp.EnableMouse(true)
	siriApp.EnablePaste(true)

	siriPage := newSiriPage(siriApp, siriClient, sendTemplates, responseTemplates)
	helpPage := newHelpPage()

	pages := tview.NewPages()
	pages.AddAndSwitchToPage("siri", siriPage, true)
	pages.AddPage("help", helpPage, true, false)

	siriApp.SetRoot(pages, true).SetFocus(pages)

	if len(siriApp.focusComponents) > 0 {
		siriApp.SetFocus(siriApp.focusComponents[0])
	}

	// Installing shortcuts
	siriApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlX:
			cancel(nil)
		case tcell.KeyCtrlO:
			siriPage.send()
			return nil
		case tcell.KeyCtrlC:
			return nil
		case tcell.KeyTab:
			nextFocus(siriApp)
		case tcell.KeyBacktab:
			prevFocus(siriApp)
		case tcell.KeyF1:
			if pages.GetPage("siri").HasFocus() {
				pages.SwitchToPage("help")
			} else {
				pages.SwitchToPage("siri")
			}
		}
		return event
	})

	return siriApp
}

func (app *SiriApp) register(prioritizedComponents ...tview.Primitive) {
	app.focusComponents = append(app.focusComponents, prioritizedComponents...)
}

func nextFocus(app *SiriApp) {
	switchFocus(app, 1)
}

func prevFocus(app *SiriApp) {
	switchFocus(app, -1)
}

func switchFocus(app *SiriApp, direction int) {
	focusElementsCount := len(app.focusComponents)
	if focusElementsCount == 0 {
		return
	}
	currentFocus := app.GetFocus()
	for i, component := range app.focusComponents {
		if component == currentFocus {
			nextFocus := app.focusComponents[(i+direction+focusElementsCount)%focusElementsCount]
			app.SetFocus(nextFocus)
			return
		}
	}
	app.SetFocus(app.focusComponents[0])
}
