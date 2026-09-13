package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"common-svr/internal/bootstrap"
	"common-svr/internal/common/config"
	commonlogger "common-svr/internal/common/logger"
)

func main() {
	if err := run(); err != nil {
		slog.Error("common-svr stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := commonlogger.New(cfg.Log.Level).With(
		"service", cfg.App.Name,
		"environment", cfg.App.Environment,
	)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := bootstrap.New(ctx, cfg, logger)
	if err != nil {
		return fmt.Errorf("bootstrap application: %w", err)
	}
	defer app.Close()

	logger.Info("starting common-svr", "address", cfg.Server.Address())
	return app.Run(ctx)
}
