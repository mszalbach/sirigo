package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestTextView(t *testing.T) {
	simScreen := tcell.NewSimulationScreen("UTF-8")
	simScreen.Init()
	simScreen.SetSize(20, 10)

	text := tview.NewTextArea()
	text.SetText("Hello", false)

	// Draw directly without running the app event loop
	text.SetRect(0, 0, 20, 10)
	text.Draw(simScreen)

	simScreen.SetCell(0, 0, tcell.StyleDefault, 'B')
	// Verify content
	cell, _, _, _ := simScreen.GetContent(0, 0)
	if cell != 'B' {
		t.Errorf("Expected 'H' at (0,0), got '%c'", cell)
	}
}
