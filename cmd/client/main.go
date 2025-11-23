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
	file, err := os.OpenFile(cfg.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		panic(err)
	}
	logger := slog.New(slog.NewJSONHandler(file, nil))
	defer file.Close()
	slog.SetDefault(logger)

	siriClient := siri.NewClient(cfg.clientPort)
	tc := siri.NewTemplateCache(cfg.templateDir)
	templateNames, templateErr := tc.TemplateNames()

	ui.InitStyles()

	controlBar := tview.NewTextView()
	controlBar.SetDynamicColors(true)
	// TODO use style colors
	controlBar.SetText("[black:white]^O[white:#282a36] Send [black:white]^X[white:#282a36] Exit")

	statusBar := tview.NewTextView()
	statusBar.SetDynamicColors(true)
	statusBar.SetText("[red]Error")

	serverResonseView := tview.NewTextView()
	serverResonseView.SetDynamicColors(true).SetBorder(true).SetTitle("Server Response")

	serverRequestView := tview.NewTextView()
	serverRequestView.SetDynamicColors(true).SetBorder(true).SetTitle("Server Request")

	urlInput := tview.NewInputField().SetPlaceholder("http://localhost:8080").SetText(cfg.url)
	urlInput.SetLabel("URL: ")
	urlInput.SetFieldWidth(40)

	bodyInput := tview.NewTextArea()
	bodyInput.SetBorder(true).SetTitle("Client Request")
	app := tview.NewApplication()
	app.EnableMouse(true)
	app.EnablePaste(true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlX:
			app.Stop()
			// not the correct way to shutdown app and webserver, but works for now
			os.Exit(0)
		case tcell.KeyCtrlO:
			response, sendErr := siriClient.Send(urlInput.GetText(), "")

			if sendErr != nil {
				statusBar.SetText("[red]" + sendErr.Error())
				return nil
			}
			serverResonseView.SetText(tview.TranslateANSI(ui.Highlight(response.Body, response.Language)))
			return nil
		case tcell.KeyCtrlC:
			return nil
		}
		return event
	})

	dropdown := tview.NewDropDown().SetLabel("Templates: ")

	if templateErr == nil {
		dropdown.SetOptions(templateNames, nil)
	} else {
		statusBar.SetText("[error]" + templateErr.Error())
	}

	dropdown.SetSelectedFunc(func(text string, index int) {
		et, err := tc.ExecuteTemplate(text, siri.Data{Now: time.Now(), ClientRef: cfg.clientRef})
		if err != nil {
			statusBar.SetText("[red]" + err.Error())
			return
		}
		bodyInput.SetText(et, false)
	})

	autoresponse := tview.NewDropDown().SetLabel("Client auto response: ")
	autoresponse.SetOptions([]string{"aaa"}, nil)
	autoresponse.SetCurrentOption(0)

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
		AddItem(controlBar, 0, 1, false).AddItem(statusBar, 0, 1, false)
	appFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(bodyFlex, 0, 1, false).
		AddItem(footerFlex, 2, 0, false)

	// TODO: the GUI can end and the server will still run
	var g errgroup.Group

	app.SetRoot(appFlex, true).SetFocus(urlInput)
	g.Go(siriClient.ListenAndServe)
	g.Go(app.Run)

	panic(g.Wait())
}
