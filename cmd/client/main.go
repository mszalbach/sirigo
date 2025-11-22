package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/mszalbach/sirigo/internal/ui"
	"github.com/rivo/tview"
	"golang.org/x/sync/errgroup"
)

func main() {

	cfg := loadConfig()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	file, err := os.OpenFile(cfg.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	slog.SetDefault(logger)

	siriClient := siri.NewClient(cfg.clientPort)
	ui.InitStyles()

	tc := siri.NewTemplateCache(cfg.templateDir)

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
	// TODO: free up ctrl+c for ctrl+q and implement copy functionality
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

	// TODO: the GUI can end and the server will still run. Not sure if this is a problem?
	var g errgroup.Group

	app.SetRoot(appFlex, true).SetFocus(urlInput)
	g.Go(func() error { return siriClient.ListenAndServe() })
	g.Go(func() error { return app.Run() })

	panic(g.Wait())
}
