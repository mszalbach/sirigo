// Package main starts the TUI and everything needed to provide a SIRI client
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
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
	clientTemplates, err := siri.NewTemplateCache(cfg.templateDir)
	if err != nil {
		panic(err)
	}
	serverTemplates, err := siri.NewTemplateCache(cfg.autoresponseDir)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	app := ui.NewSiriApp(siriClient, clientTemplates, serverTemplates, cancel)

	eg, _ := errgroup.WithContext(ctx)
	// TODO on "port is already used" the error is not returned directly
	eg.Go(siriClient.ListenAndServe)
	eg.Go(app.Run)
	eg.Go(func() error {
		<-ctx.Done()
		app.Stop()
		return siriClient.Stop(ctx)
	})

	werr := eg.Wait()
	if errors.Is(werr, http.ErrServerClosed) {
		slog.Info("Application closed", slog.Any("context", werr))
	} else {
		slog.Error("Something unexpected closed the app", slog.Any("error", werr))
		os.Exit(1)
	}
}
