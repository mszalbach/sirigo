package ui

import (
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/rivo/tview"
)

type siriApp struct {
	*tview.Application
	siriClient        siri.Client
	sendTemplates     siri.TemplateCache
	responseTemplates siri.TemplateCache
}

func NewSiriApp(
	siriClient siri.Client,
	sendTemplates siri.TemplateCache,
	responseTemplates siri.TemplateCache,
) siriApp {
	app := siriApp{
		Application:       tview.NewApplication(),
		siriClient:        siriClient,
		sendTemplates:     sendTemplates,
		responseTemplates: responseTemplates,
	}

	initStyles()
	app.EnableMouse(true)
	app.EnablePaste(true)

	// building ui elements
	errorChannel := make(chan error, 5)
	statusBar := newStatusBar(errorChannel)
	keymap := newKeymap()
	sendView := newSiriClientView(siriClient, sendTemplates, errorChannel)

	// TODO cleanup into own go files
	serverResonseView := tview.NewTextView()
	serverResonseView.SetDynamicColors(true).SetBorder(true).SetTitle("Server Response")

	serverRequestView := tview.NewTextView()
	serverRequestView.SetDynamicColors(true).SetBorder(true).SetTitle("Server Request")

	autoresponse := tview.NewDropDown().SetLabel("Client auto response: ")
	autoresponse.SetOptions([]string{"aaa"}, nil)
	autoresponse.SetCurrentOption(0)

	// building layout
	reveiveFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(autoresponse, 2, 0, false).
		AddItem(serverResonseView, 0, 2, false).
		AddItem(serverRequestView, 0, 1, false)

	bodyFlex := tview.NewFlex().
		AddItem(sendView, 0, 1, false).
		AddItem(reveiveFlex, 0, 1, false)

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
			response := sendView.send()
			serverResonseView.SetText(tview.TranslateANSI(highlight(response.Body, response.Language)))
			return nil
		case tcell.KeyCtrlC:
			return nil
		}
		return event
	})

	return app
}
