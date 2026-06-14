package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"roboticCrewChallenge/internal/config"
	"roboticCrewChallenge/internal/platform/logging"
	"roboticCrewChallenge/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := logging.New(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := server.New(cfg, logger)
	return srv.Run(ctx)
}
