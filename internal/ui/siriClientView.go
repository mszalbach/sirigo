package ui

import (
	"fmt"

	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/rivo/tview"
)

type siriClientView struct {
	*tview.Flex
	siriClient   siri.Client
	errorChannel chan<- error
	urlInput     *tview.InputField
	requestArea  *tview.TextArea
}

func newSiriClientView(
	siriClient siri.Client,
	sendTemplates siri.TemplateCache,
	errorChannel chan<- error,
) siriClientView {
	urlInput := tview.NewInputField().SetPlaceholder("http://localhost:8080")
	urlInput.SetLabel("URL: ")
	urlInput.SetFieldWidth(80)
	urlInput.SetText(siriClient.ServerURL)

	siriClientRequestArea := tview.NewTextArea()
	siriClientRequestArea.SetBorder(true).SetTitle(fmt.Sprintf("Client Request (clientRef: %s)", siriClient.ClientRef))

	dropdown := tview.NewDropDown().SetLabel("Templates: ")

	templateNames, err := sendTemplates.TemplateNames()
	if err == nil {
		dropdown.SetOptions(templateNames, nil)
	} else {
		errorChannel <- err
	}

	dropdown.SetSelectedFunc(func(name string, _ int) {
		requestTemplate, err := sendTemplates.GetTemplate(name)
		if err != nil {
			errorChannel <- err
			return
		}
		urlPath := siri.GetURLPathFromTemplate(requestTemplate)
		urlInput.SetText(siriClient.ServerURL + urlPath)
		siriClientRequestArea.SetText(requestTemplate, false)
	})

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(urlInput, 2, 0, true).
		AddItem(dropdown, 2, 0, false).
		AddItem(siriClientRequestArea, 0, 1, false)

	return siriClientView{
		Flex:         flex,
		siriClient:   siriClient,
		errorChannel: errorChannel,
		urlInput:     urlInput,
		requestArea:  siriClientRequestArea,
	}
}

func (sc siriClientView) send() siri.ServerResponse {
	res, err := sc.siriClient.Send(sc.urlInput.GetText(), sc.requestArea.GetText())
	if err != nil {
		sc.errorChannel <- err
	}
	return res
}
