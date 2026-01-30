// Package main starts the TUI and everything needed to provide a SIRI client
package main

import (
	"context"
	"errors"
	"fmt"
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
	logFile, err := os.OpenFile(cfg.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		panic(err)
	}
	logger := slog.New(slog.NewJSONHandler(logFile, nil))
	defer logFile.Close()
	slog.SetDefault(logger)

	cancelContext, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	stopContext, stop := signal.NotifyContext(cancelContext, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	httpLogFile, err := os.OpenFile(cfg.httpLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		panic(err)
	}
	defer httpLogFile.Close()
	siriClient := siri.NewClient(cfg.clientRef, cfg.url, cfg.clientPort, httpLogFile)

	clientTemplates, err := siri.NewTemplateCache(cfg.templateDir)
	if err != nil {
		panic(err)
	}
	serverTemplates, err := siri.NewTemplateCache(cfg.autoresponseDir)
	if err != nil {
		panic(err)
	}

	app := ui.NewSiriApp(&siriClient, clientTemplates, serverTemplates, cancel)

	go func() {
		if err := app.Run(); err != nil {
			slog.Error("App could not be started", slog.Any("error", err))
			cancel(err)
		}
	}()
	go func() {
		if err := siriClient.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error(
				"SIRI client could not be started",
				slog.String("address", cfg.clientPort),
				slog.Any("error", err),
			)
			cancel(err)
		}
	}()

	<-stopContext.Done()
	slog.Info("Graceful shutdown")
	app.Stop()

	timeoutCtx, timeoutFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer timeoutFunc()
	if stopErr := siriClient.Stop(timeoutCtx); stopErr != nil {
		slog.Warn("server stop failed", slog.Any("error", stopErr))
	}

	if err := context.Cause(stopContext); !errors.Is(err, context.Canceled) {
		slog.Error("App could not be started: ", slog.Any("error", err))
		fmt.Println("App could not be started:", err)
	}
}
