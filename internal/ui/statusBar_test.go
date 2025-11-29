package ui

import (
	"errors"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestScreen(t *testing.T) tcell.SimulationScreen {
	screen := tcell.NewSimulationScreen("")
	err := screen.Init()
	require.NoError(t, err)
	return screen
}

func Test_is_empty_without_channel(t *testing.T) {
	// Given
	screen := newTestScreen(t)
	defer screen.Fini()
	box := newStatusBar(nil)

	// When
	box.Draw(screen)

	// Then
	actualLine := getScreenTextLine(screen, 0, 10)
	assert.Empty(t, actualLine)
}

func Test_shows_error_when_one_is_sent_via_channel(t *testing.T) {
	// Given
	screen := newTestScreen(t)
	defer screen.Fini()
	channel := make(chan error)

	box := newStatusBar(channel)

	// When
	channel <- errors.New("Failed to send")
	box.Draw(screen)

	// Then
	actualLine := getScreenTextLine(screen, 0, 20)
	assert.Equal(t, "Failed to send", actualLine)
}

func getScreenTextLine(screen tcell.SimulationScreen, y int, length int) string {
	var builder strings.Builder
	for x := range length {
		content, _, _ := screen.Get(x, y)
		builder.WriteString(content)
	}
	return strings.TrimSpace(builder.String())
}
