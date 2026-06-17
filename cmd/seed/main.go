// Command seed creates demo accounts (a merchant with a store, and a customer)
// so the deployed system can be exercised through the authenticated GraphQL API
// without an out-of-band SQL insert — emails are encrypted and blind-indexed, so
// accounts can only be created through the same code path the app uses. It reads
// DATABASE_URL and PII_ENCRYPTION_KEY from the environment and is idempotent:
// re-running it leaves existing demo accounts untouched.
package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"

	"roboticCrewChallenge/internal/adapter/postgres"
	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/platform/crypto"
)

const (
	merchantEmail = "merchant@petstore.local"
	customerEmail = "customer@petstore.local"
	demoPassword  = "demo-password"
	storeName     = "Demo Store"
)

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	encodedKey := os.Getenv("PII_ENCRYPTION_KEY")
	if encodedKey == "" {
		return errors.New("PII_ENCRYPTION_KEY is required")
	}
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return fmt.Errorf("decode PII_ENCRYPTION_KEY: %w", err)
	}

	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	encryptor, err := crypto.NewEncryptor(key)
	if err != nil {
		return fmt.Errorf("init encryptor: %w", err)
	}
	blindIndex, err := crypto.NewBlindIndex(key)
	if err != nil {
		return fmt.Errorf("init blind index: %w", err)
	}

	merchants := postgres.NewMerchantRepository(pool, encryptor, blindIndex)
	stores := postgres.NewStoreRepository(pool)
	customers := postgres.NewCustomerRepository(pool, encryptor, blindIndex)

	storeID, err := seedMerchantWithStore(ctx, merchants, stores)
	if err != nil {
		return err
	}
	if err := seedCustomer(ctx, customers); err != nil {
		return err
	}

	slog.Info("demo accounts ready",
		"merchant", merchantEmail,
		"customer", customerEmail,
		"storeId", storeID,
	)
	return nil
}

func seedMerchantWithStore(ctx context.Context, merchants *postgres.MerchantRepository, stores *postgres.StoreRepository) (uuid.UUID, error) {
	hash, err := crypto.HashPassword(demoPassword)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash merchant password: %w", err)
	}
	merchant, err := domain.NewMerchant(uuid.New(), merchantEmail, hash, time.Now().UTC())
	if err != nil {
		return uuid.Nil, err
	}

	merchantID := merchant.ID
	if err := merchants.Create(ctx, merchant); err != nil {
		if !errors.Is(err, domain.ErrEmailInUse) {
			return uuid.Nil, fmt.Errorf("create merchant: %w", err)
		}
		existing, err := merchants.GetByEmail(ctx, merchantEmail)
		if err != nil {
			return uuid.Nil, fmt.Errorf("look up existing merchant: %w", err)
		}
		merchantID = existing.ID
	}

	store, err := domain.NewStore(uuid.New(), merchantID, storeName, time.Now().UTC())
	if err != nil {
		return uuid.Nil, err
	}
	if err := stores.Create(ctx, store); err != nil {
		if !errors.Is(err, domain.ErrStoreAlreadyExists) {
			return uuid.Nil, fmt.Errorf("create store: %w", err)
		}
		existing, err := stores.GetByMerchantID(ctx, merchantID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("look up existing store: %w", err)
		}
		return existing.ID, nil
	}
	return store.ID, nil
}

func seedCustomer(ctx context.Context, customers *postgres.CustomerRepository) error {
	hash, err := crypto.HashPassword(demoPassword)
	if err != nil {
		return fmt.Errorf("hash customer password: %w", err)
	}
	customer, err := domain.NewCustomer(uuid.New(), customerEmail, hash, time.Now().UTC())
	if err != nil {
		return err
	}
	if err := customers.Create(ctx, customer); err != nil && !errors.Is(err, domain.ErrEmailInUse) {
		return fmt.Errorf("create customer: %w", err)
	}
	return nil
}
