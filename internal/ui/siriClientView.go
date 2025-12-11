package ui

import (
	"fmt"

	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/rivo/tview"
)

type siriClientView struct {
	*tview.Flex
	model        *siriClientViewModel
	siriClient   siri.Client
	errorChannel chan<- error
}

type siriClientViewModel struct {
	body string
	url  string
}

func newSiriClientView(
	siriClient siri.Client,
	sendTemplates siri.TemplateCache,
	errorChannel chan<- error,
) siriClientView {
	model := siriClientViewModel{
		body: "",
		url:  "",
	}

	urlInput := tview.NewInputField().SetPlaceholder("http://localhost:8080")
	urlInput.SetLabel("URL: ")
	urlInput.SetFieldWidth(80)
	urlInput.SetText(siriClient.ServerURL)
	urlInput.SetChangedFunc(func(url string) {
		model.url = url
	})

	siriClientRequestArea := tview.NewTextArea()
	siriClientRequestArea.SetBorder(true).SetTitle(fmt.Sprintf("Client Request (clientRef: %s)", siriClient.ClientRef))
	siriClientRequestArea.SetChangedFunc(func() {
		model.body = siriClientRequestArea.GetText()
	})

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
		// If no server URL is configured, let the user enter the full URL
		if siriClient.ServerURL != "" {
			urlInput.SetText(siriClient.ServerURL + urlPath)
		}
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
		model:        &model,
		errorChannel: errorChannel,
	}
}

func (sc siriClientView) send() siri.ServerResponse {
	res, err := sc.siriClient.Send(sc.model.url, sc.model.body)
	if err != nil {
		sc.errorChannel <- err
	}
	return res
}
