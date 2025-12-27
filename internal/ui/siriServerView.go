package ui

import (
	"fmt"

	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/rivo/tview"
)

type siriServerView struct {
	*tview.Flex
	serverResponseTextView *codeTextView
}

func newSiriServerView(
	app tuiApp,
	siriClient *siri.Client,
	responseTemplates siri.TemplateCache,
	errorChannel chan<- error,
) siriServerView {
	serverResponseTextView := newCodeTextView(app, "Server Response")
	serverRequestTextView := newCodeTextView(app, "Server Request")
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

	go listenForServerRequests(serverRequestTextView, siriClient)

	// register focus order
	app.register(autoresponseDropdown, serverResponseTextView, serverRequestTextView)

	return siriServerView{
		Flex:                   siriServerFlex,
		serverResponseTextView: serverResponseTextView,
	}
}

func listenForServerRequests(serverRequestTextView *codeTextView, siriClient *siri.Client) {
	for req := range siriClient.ServerRequest {
		body := fmt.Sprintf("<!-- %s%s -->\n%s", req.RemoteAddress, req.URL, req.Body)
		serverRequestTextView.SetCode(body, req.Language)
	}
}

func (sv siriServerView) setResponse(response siri.ServerResponse) {
	sv.serverResponseTextView.SetCode(response.Body, response.Language)
}
