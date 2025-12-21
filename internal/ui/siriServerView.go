package ui

import (
	"fmt"

	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/rivo/tview"
)

type siriServerView struct {
	app tuiApp
	*tview.Flex
	serverResponseTextView *tview.TextView
}

func newSiriServerView(
	app tuiApp,
	siriClient *siri.Client,
	responseTemplates siri.TemplateCache,
	errorChannel chan<- error,
) siriServerView {
	serverResponseTextView := tview.NewTextView()
	serverResponseTextView.SetDynamicColors(true).SetBorder(true).SetTitle("Server Response")

	serverRequestTextView := tview.NewTextView()
	serverRequestTextView.SetDynamicColors(true).SetBorder(true).SetTitle("Server Request")

	autoresponseDropdown := tview.NewDropDown().SetLabel("Client auto-response: ")

	templateNames, err := responseTemplates.TemplateNames()
	if err == nil {
		autoresponseDropdown.SetOptions(templateNames, nil)
	} else {
		errorChannel <- err
	}

	autoresponseDropdown.SetOptions(templateNames, nil)
	autoresponseDropdown.SetSelectedFunc(func(name string, _ int) {
		template, err := responseTemplates.GetTemplate(name)
		if err != nil {
			errorChannel <- err
			return
		}
		siriClient.AutoClientResponse.Body = template
	})
	autoresponseDropdown.SetCurrentOption(0)

	siriServerFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(autoresponseDropdown, 2, 0, false).
		AddItem(serverResponseTextView, 0, 2, false).
		AddItem(serverRequestTextView, 0, 1, false)

	go listenForServerRequests(app, serverRequestTextView, siriClient)

	// register focus order
	app.register(autoresponseDropdown, serverResponseTextView, serverRequestTextView)

	return siriServerView{
		Flex:                   siriServerFlex,
		serverResponseTextView: serverResponseTextView,
		app:                    app,
	}
}

func listenForServerRequests(app tuiApp, serverRequestTextView *tview.TextView, siriClient *siri.Client) {
	for req := range siriClient.ServerRequest {
		body := fmt.Sprintf("<!-- %s%s -->\n%s", req.RemoteAddress, req.URL, req.Body)
		serverRequestTextView.ScrollToBeginning()
		// changing UI from an async go routine, so it needs to use the queueUpdate methods
		app.QueueUpdateDraw(func() {
			serverRequestTextView.SetText(tview.TranslateANSI(highlight(body, req.Language)))
		})
	}
}

func (sv siriServerView) setResponse(response siri.ServerResponse) {
	sv.serverResponseTextView.ScrollToBeginning()
	sv.serverResponseTextView.SetText(response.Body)
	// highlight takes a lot of time for big responses, so doing it delayed later
	go func() {
		highlighted := tview.TranslateANSI(highlight(response.Body, response.Language))
		sv.app.QueueUpdateDraw(func() {
			sv.serverResponseTextView.SetText(highlighted)
		})
	}()
}
