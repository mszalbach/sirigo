package ui

import "github.com/rivo/tview"

type helpPage struct {
	*tview.Flex
}

func newHelpPage() *helpPage {
	helpPage := &helpPage{
		Flex: tview.NewFlex(),
	}
	textview := tview.NewTextView()
	textview.SetBorder(true)
	textview.SetTitle("Help")
	textview.SetDynamicColors(true)
	// TODO use colors from style.go
	textview.SetText(`Sirigo is designed to be a SIRI client to send and receive SIRI messages.

Global Keybindings:

F1: 	Show this help page / Close this help page
Ctrl-X: Exit the application

SIRI page Keybindings:

Ctrl-O: 	   Send a SIRI request
Tab/Shift+Tab: Cycle focus between components

Client Request:

Ctrl-D:		   Delete the character under the cursor (or the first character on the next line if the cursor is at the end of a line).
Alt-Backspace: Delete the word to the left of the cursor.
Ctrl-K:        Delete everything under and to the right of the cursor until the next newline character.
Ctrl-W:        Delete from the start of the current word to the left of the cursor.
Ctrl-U:        Delete the current line, i.e. everything after the last newline character before the cursor up until the next newline character. This may span multiple visible rows if wrapping is enabled.

Server Response / Server Request:

h: 		Move left.
l: 		Move right.
j: 		Move down.
k: 		Move up.
g: 		Move to the top.
G: 		Move to the bottom.
Ctrl-F: Move down by one page.
Ctrl-B: Move up by one page.
`)
	helpPage.AddItem(textview, 0, 1, true)
	return helpPage
}
