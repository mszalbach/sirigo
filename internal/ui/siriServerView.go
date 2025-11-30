package ui

import (
	"fmt"

	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/rivo/tview"
)

type siriServerView struct {
	*tview.Flex
	serverResponseTextView *tview.TextView
}

func newSiriServerView(
	siriClient siri.Client,
	responseTemplates siri.TemplateCache,
	errorChannel chan<- error,
) siriServerView {
	serverResponseTextView := tview.NewTextView()
	serverResponseTextView.SetDynamicColors(true).SetBorder(true).SetTitle("Server Response")

	serverRequestTextView := tview.NewTextView()
	serverRequestTextView.SetDynamicColors(true).SetBorder(true).SetTitle("Server Request")

	autoresponseDropdown := tview.NewDropDown().SetLabel("Client auto-response: ")

	templateNames, templateErr := responseTemplates.TemplateNames()
	if templateErr == nil {
		autoresponseDropdown.SetOptions(templateNames, nil)
	} else {
		errorChannel <- templateErr
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

	go listenForServerRequests(siriClient, serverRequestTextView)

	return siriServerView{
		Flex:                   siriServerFlex,
		serverResponseTextView: serverResponseTextView,
	}
}

func listenForServerRequests(siriClient siri.Client, serverRequestTextView *tview.TextView) {
	for req := range siriClient.ServerRequest {
		body := fmt.Sprintf("<!-- %s%s -->\n%s", req.RemoteAddress, req.URL, req.Body)
		serverRequestTextView.ScrollToBeginning()
		serverRequestTextView.SetText(tview.TranslateANSI(highlight(body, req.Language)))
	}
}

func (sv siriServerView) setResponse(response siri.ServerResponse) {
	sv.serverResponseTextView.ScrollToBeginning()
	sv.serverResponseTextView.SetText(tview.TranslateANSI(highlight(response.Body, response.Language)))
}
