package ui

import (
	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/rivo/tview"
)

type siriPage struct {
	*tview.Flex
	name           string
	siriClientView siriClientView
	siriServerView siriServerView
	statusBar      statusBar
}

func newSiriPage(siriApp tuiApp, siriClient *siri.Client,
	sendTemplates siri.TemplateCache,
	responseTemplates siri.TemplateCache,
) *siriPage {
	siriPage := siriPage{
		name: "siri",
		Flex: tview.NewFlex(),
	}

	// Building UI elements
	errorChannel := make(chan error, 5)
	siriPage.statusBar = newStatusBar(siriApp, errorChannel)
	keymap := newKeymap()
	siriPage.siriClientView = newSiriClientView(siriApp, siriClient, sendTemplates, errorChannel)
	siriPage.siriServerView = newSiriServerView(siriApp, siriClient, responseTemplates, errorChannel)

	// Building layout
	bodyFlex := tview.NewFlex().
		AddItem(siriPage.siriClientView, 0, 1, false).
		AddItem(siriPage.siriServerView, 0, 1, false)

	footerFlex := tview.NewFlex().
		AddItem(keymap, 0, 1, false).AddItem(siriPage.statusBar, 0, 1, false)

	siriPage.Flex.
		SetDirection(tview.FlexRow).
		AddItem(bodyFlex, 0, 1, false).
		AddItem(footerFlex, 2, 0, false)

	return &siriPage
}

func (sp *siriPage) send() {
	// async since request can take some time and block the UI
	// better to inform the user about it
	go func() {
		response := sp.siriClientView.send()
		sp.siriServerView.setResponse(response)
	}()
}
