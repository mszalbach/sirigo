package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/mszalbach/sirigo/internal/ui"
	"github.com/rivo/tview"
	"golang.org/x/sync/errgroup"
)

type Data struct {
	Now       string
	ClientRef string
}

func main() {
	cfg := loadConfig()

	siriClient := siri.NewClient(cfg.clientPort)
	ui.InitStyles()

	tc := siri.NewTemplateCache(cfg.templateDir)

	go func() {
		for {
			val := <-siriClient.ServerRequest
			fmt.Println("GORoutine 1", val)
		}

	}()

	responseView := tview.NewTextView()
	responseView.SetDynamicColors(true).SetBorder(true).SetTitle("Response Body")

	urlInput := tview.NewInputField().SetPlaceholder("https://www.w3schools.com/xml/note.xml").SetText(cfg.url)
	urlInput.SetLabel("URL: ")
	urlInput.SetFieldWidth(40)
	urlInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			response := siriClient.Send(urlInput.GetText(), "")
			responseView.SetText(tview.TranslateANSI(ui.Highlight(response.Body, response.Language)))
			return nil
		}
		return event
	})

	bodyInput := tview.NewTextArea()
	bodyInput.SetBorder(true).SetTitle("Request Body")
	app := tview.NewApplication()
	app.EnableMouse(true)
	// TODO ctrl+c frei machen für strg+q oder sowas? und ctrl+c dann selber als Copy umsetzen
	app.EnablePaste(true)

	dropdown := tview.NewDropDown().
		SetLabel("Templates: ").
		SetOptions(tc.TemplateNames(), nil)
	dropdown.SetSelectedFunc(func(text string, index int) {
		et := tc.ExecuteTemplate(text, siri.Data{Now: time.Now(), ClientRef: cfg.clientRef})
		bodyInput.SetText(et, false)
	})

	sendFlex := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(urlInput, 2, 0, true).AddItem(dropdown, 2, 0, false).AddItem(bodyInput, 0, 1, false)
	appFlex := tview.NewFlex().AddItem(sendFlex, 0, 1, false).AddItem(responseView, 0, 1, false)

	//TODO the gui can end and the server will still run. Not sure if this is a problem?
	var g errgroup.Group

	app.SetRoot(appFlex, true).SetFocus(urlInput)
	g.Go(func() error { return siriClient.ListenAndServer() })
	g.Go(func() error { return app.Run() })

	panic(g.Wait())
}
