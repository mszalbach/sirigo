// Package main starts the TUI and everything needed to provide a SIRI client
package main

import (
	"log/slog"
	"os"

	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/mszalbach/sirigo/internal/ui"
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

	siriClient := siri.NewClient(cfg.clientRef, cfg.url, cfg.clientPort)
	tc := siri.NewTemplateCache(cfg.templateDir)

	app := ui.NewSiriApp(siriClient, tc, tc)

	// TODO: the GUI can end and the server will still run
	var g errgroup.Group
	g.Go(siriClient.ListenAndServe)
	g.Go(app.Run)

	panic(g.Wait())
}
