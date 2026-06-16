package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"roboticCrewChallenge/internal/adapter/objectstore"
	"roboticCrewChallenge/internal/adapter/postgres"
	"roboticCrewChallenge/internal/auth"
	"roboticCrewChallenge/internal/config"
	"roboticCrewChallenge/internal/platform/crypto"
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

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	pictureStore, err := objectstore.New(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOBucket, cfg.MinIOUseSSL)
	if err != nil {
		return err
	}
	if err := pictureStore.EnsureBucket(ctx); err != nil {
		return err
	}

	encryptor, err := crypto.NewEncryptor(cfg.PIIEncryptionKey)
	if err != nil {
		return err
	}
	blindIndex, err := crypto.NewBlindIndex(cfg.PIIEncryptionKey)
	if err != nil {
		return err
	}

	authenticator := auth.NewAuthenticator(
		postgres.NewMerchantRepository(pool, encryptor, blindIndex),
		postgres.NewCustomerRepository(pool, encryptor, blindIndex),
		postgres.NewStoreRepository(pool),
	)

	srv := server.New(cfg, logger, authenticator)
	return srv.Run(ctx)
}
