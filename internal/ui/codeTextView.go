package ui

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"runtime"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type codeTextView struct {
	*tview.TextView
	app tuiApp
}

func newCodeTextView(app tuiApp, title string) *codeTextView {
	codeTextView := &codeTextView{tview.NewTextView(), app}
	codeTextView.SetDynamicColors(true).SetBorder(true).SetTitle(title)

	codeTextView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlE {
			app.Suspend(codeTextView.openInEditor)
			return nil
		}
		return event
	})
	return codeTextView
}

func (ctv *codeTextView) SetCode(code string, language string) {
	ctv.ScrollToBeginning()
	ctv.SetText(code)
	// highlight takes a lot of time for big responses, so doing it delayed later
	go func() {
		highlighted := tview.TranslateANSI(highlight(code, language))
		ctv.app.QueueUpdateDraw(func() {
			ctv.SetText(highlighted)
		})
	}()
}

func (ctv *codeTextView) openInEditor() {
	f, err := os.CreateTemp("", ctv.GetTitle()+"-*.txt")
	if err != nil {
		return
	}
	defer os.Remove(f.Name())

	_, err = f.WriteString(ctv.GetText(true))
	if err != nil {
		slog.Warn("Could not write to tmp file", slog.String("file", f.Name()), slog.Any("error", err))
		return
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			editor = "notepad.exe"
		} else {
			editor = "vi"
		}
	}

	cmd := exec.CommandContext( // nolint:gosec // when some one captures the EDITOR env you have bigger problems
		context.Background(),
		editor,
		f.Name(),
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		slog.Warn(
			"Could not start editor",
			slog.String("editor", editor),
			slog.String("file", f.Name()),
			slog.Any("error", err),
		)
	}
}
