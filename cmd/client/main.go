// Package main starts the TUI and everything needed to provide a SIRI client
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/mszalbach/sirigo/internal/ui"
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

	cancelContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	stopContext, stop := signal.NotifyContext(cancelContext, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	siriClient := siri.NewClient(cfg.clientRef, cfg.url, cfg.clientPort)

	clientTemplates, err := siri.NewTemplateCache(cfg.templateDir)
	if err != nil {
		panic(err)
	}
	serverTemplates, err := siri.NewTemplateCache(cfg.autoresponseDir)
	if err != nil {
		panic(err)
	}

	app := ui.NewSiriApp(siriClient, clientTemplates, serverTemplates, cancel)

	go func() {
		if err := app.Run(); err != nil {
			slog.Error("App could not be started", slog.Any("error", err))
			panic("App could not be started")
		}
	}()
	go func() {
		if err := siriClient.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error(
				"SIRI client could not be started",
				slog.String("address", cfg.clientPort),
				slog.Any("error", err),
			)
			panic("Server not working")
		}
	}()

	<-stopContext.Done()
	slog.Info("Graceful shutdown")
	app.Stop()

	timeoutCtx, timeoutFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer timeoutFunc()
	if stopErr := siriClient.Stop(timeoutCtx); stopErr != nil {
		slog.Warn("server stop failed", slog.Any("error", stopErr.Error()))
	}
}
