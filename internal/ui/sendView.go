package ui

import (
	"time"

	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/rivo/tview"
)

type sendView struct {
	*tview.Flex
	model        *sendViewModel
	siriClient   siri.Client
	errorChannel chan<- error
}

type sendViewModel struct {
	body string
	url  string
}

func newSendView(siriClient siri.Client, sendTemplates siri.TemplateCache, errorChannel chan<- error) sendView {
	model := sendViewModel{
		body: "",
		url:  "",
	}

	urlInput := tview.NewInputField().SetPlaceholder("http://localhost:8080")
	urlInput.SetLabel("URL: ")
	urlInput.SetFieldWidth(40)
	urlInput.SetText(siriClient.ServerURL)
	urlInput.SetChangedFunc(func(url string) {
		model.url = url
	})

	bodyInput := tview.NewTextArea()
	bodyInput.SetBorder(true).SetTitle("Client Request")
	bodyInput.SetChangedFunc(func() {
		model.body = bodyInput.GetText()
	})

	dropdown := tview.NewDropDown().SetLabel("Templates: ")

	templateNames, templateErr := sendTemplates.TemplateNames()
	if templateErr == nil {
		dropdown.SetOptions(templateNames, nil)
	} else {
		errorChannel <- templateErr
	}

	dropdown.SetSelectedFunc(func(text string, _ int) {
		et, err := sendTemplates.ExecuteTemplate(text, siri.Data{Now: time.Now(), ClientRef: siriClient.ClientRef})
		if err != nil {
			errorChannel <- err
			return
		}
		bodyInput.SetText(et, false)
	})

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(urlInput, 2, 0, true).
		AddItem(dropdown, 2, 0, false).
		AddItem(bodyInput, 0, 1, false)

	return sendView{
		Flex:         flex,
		siriClient:   siriClient,
		model:        &model,
		errorChannel: errorChannel,
	}
}

func (sv sendView) send() siri.ServerResponse {
	res, err := sv.siriClient.Send(sv.model.url, sv.model.body)
	if err != nil {
		sv.errorChannel <- err
	}
	return res
}
