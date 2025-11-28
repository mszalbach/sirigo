package ui

import (
	"os"
	"time"

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

// TODO does the template caches be a problem of the client or a Presenter/Controller and not the UI?
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

	serverResonseView := tview.NewTextView()
	serverResonseView.SetDynamicColors(true).SetBorder(true).SetTitle("Server Response")

	serverRequestView := tview.NewTextView()
	serverRequestView.SetDynamicColors(true).SetBorder(true).SetTitle("Server Request")

	urlInput := tview.NewInputField().SetPlaceholder("http://localhost:8080").SetText(siriClient.ServerURL)
	urlInput.SetLabel("URL: ")
	urlInput.SetFieldWidth(40)

	bodyInput := tview.NewTextArea()
	bodyInput.SetBorder(true).SetTitle("Client Request")

	statusBar := newStatusBar()

	dropdown := tview.NewDropDown().SetLabel("Templates: ")

	templateNames, templateErr := sendTemplates.TemplateNames()
	if templateErr == nil {
		dropdown.SetOptions(templateNames, nil)
	} else {
		statusBar.error(templateErr.Error())
	}

	dropdown.SetSelectedFunc(func(text string, _ int) {
		et, err := sendTemplates.ExecuteTemplate(text, siri.Data{Now: time.Now(), ClientRef: siriClient.ClientRef})
		if err != nil {
			statusBar.error(err.Error())
			return
		}
		bodyInput.SetText(et, false)
	})

	autoresponse := tview.NewDropDown().SetLabel("Client auto response: ")
	autoresponse.SetOptions([]string{"aaa"}, nil)
	autoresponse.SetCurrentOption(0)

	keymap := newKeymap()

	sendFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(urlInput, 2, 0, true).
		AddItem(dropdown, 2, 0, false).
		AddItem(bodyInput, 0, 1, false)

	reveiveFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(autoresponse, 2, 0, false).
		AddItem(serverResonseView, 0, 2, false).
		AddItem(serverRequestView, 0, 1, false)

	bodyFlex := tview.NewFlex().
		AddItem(sendFlex, 0, 1, false).
		AddItem(reveiveFlex, 0, 1, false)

	footerFlex := tview.NewFlex().
		AddItem(keymap, 0, 1, false).AddItem(statusBar, 0, 1, false)
	appFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(bodyFlex, 0, 1, false).
		AddItem(footerFlex, 2, 0, false)
	app.SetRoot(appFlex, true).SetFocus(urlInput)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlX:
			app.Stop()
			// not the correct way to shutdown app and webserver, but works for now
			os.Exit(0)
		case tcell.KeyCtrlO:
			response, sendErr := siriClient.Send(urlInput.GetText(), "")

			if sendErr != nil {
				statusBar.error(sendErr.Error())
				return nil
			}
			serverResonseView.SetText(tview.TranslateANSI(highlight(response.Body, response.Language)))
			return nil
		case tcell.KeyCtrlC:
			return nil
		}
		return event
	})

	return app
}
