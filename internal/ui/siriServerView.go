package ui

import (
	"fmt"
	"time"

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

	autoresponseDropdown := tview.NewDropDown().SetLabel("Client auto response: ")

	templateNames, templateErr := responseTemplates.TemplateNames()
	if templateErr == nil {
		autoresponseDropdown.SetOptions(templateNames, nil)
	} else {
		errorChannel <- templateErr
	}

	autoresponseDropdown.SetOptions(templateNames, nil)
	autoresponseDropdown.SetSelectedFunc(func(text string, _ int) {
		responseBody, err := responseTemplates.ExecuteTemplate(
			text,
			siri.Data{Now: time.Now(), ClientRef: siriClient.ClientRef},
		)
		if err != nil {
			errorChannel <- err
			return
		}
		siriClient.AutoClientResponse.Body = responseBody
	})
	autoresponseDropdown.SetCurrentOption(0)

	siriServerFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(autoresponseDropdown, 2, 0, false).
		AddItem(serverResponseTextView, 0, 2, false).
		AddItem(serverRequestTextView, 0, 1, false)

	go listenFoServerRequests(siriClient, serverRequestTextView)

	return siriServerView{
		Flex:                   siriServerFlex,
		serverResponseTextView: serverResponseTextView,
	}
}

func listenFoServerRequests(siriClient siri.Client, serverRequestTextView *tview.TextView) {
	for req := range siriClient.ServerRequest {
		body := fmt.Sprintf("<!-- %s%s -->\n%s", req.RemoteAddress, req.URL, req.Body)
		serverRequestTextView.SetText(tview.TranslateANSI(highlight(body, req.Language)))
	}
}

func (sv siriServerView) setResponse(response siri.ServerResponse) {
	sv.serverResponseTextView.SetText(tview.TranslateANSI(highlight(response.Body, response.Language)))
}
