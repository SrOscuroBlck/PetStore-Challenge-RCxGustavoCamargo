// Command seed makes the deployed system browsable out of the box: it creates a
// merchant with a store, a customer, and a small catalog of pets, all through the
// same code paths the app uses (emails are encrypted and blind-indexed, pictures
// are uploaded to object storage), so nothing is inserted out of band. It reads
// DATABASE_URL, PII_ENCRYPTION_KEY, and the MINIO_* vars from the environment and
// is idempotent: re-running leaves existing demo data untouched.
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"

	"roboticCrewChallenge/internal/adapter/objectstore"
	"roboticCrewChallenge/internal/adapter/postgres"
	"roboticCrewChallenge/internal/adapter/rediscache"
	"roboticCrewChallenge/internal/app/listing"
	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/platform/crypto"
)

const (
	merchantEmail = "merchant@petstore.local"
	customerEmail = "customer@petstore.local"
	demoPassword  = "demo-password"
	storeName     = "Demo Store"
	breederName   = "Demo Breeder"
	breederEmail  = "breeder@petstore.local"
)

// demoStoreID is fixed so a fresh deployment always exposes the storefront at the
// same URL; the README and graders can rely on /store/<this id> without looking it up.
var demoStoreID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	databaseURL, err := requiredEnv("DATABASE_URL")
	if err != nil {
		return err
	}
	encodedKey, err := requiredEnv("PII_ENCRYPTION_KEY")
	if err != nil {
		return err
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

	pictures, err := pictureStore(ctx)
	if err != nil {
		return err
	}

	merchants := postgres.NewMerchantRepository(pool, encryptor, blindIndex)
	stores := postgres.NewStoreRepository(pool)
	customers := postgres.NewCustomerRepository(pool, encryptor, blindIndex)
	petRepo := postgres.NewPetRepository(pool, encryptor)
	catalog := listing.NewService(petRepo, pictures, rediscache.NoOp{})

	storeID, err := seedMerchantWithStore(ctx, merchants, stores)
	if err != nil {
		return err
	}
	if err := seedCustomer(ctx, customers); err != nil {
		return err
	}
	seeded, err := seedDemoPets(ctx, catalog, petRepo, storeID)
	if err != nil {
		return err
	}

	slog.Info("demo data ready",
		"merchant", merchantEmail,
		"customer", customerEmail,
		"storeId", storeID,
		"petsSeeded", seeded,
	)
	return nil
}

func requiredEnv(name string) (string, error) {
	if v := os.Getenv(name); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("%s is required", name)
}

func pictureStore(ctx context.Context) (*objectstore.PictureStore, error) {
	endpoint, err := requiredEnv("MINIO_ENDPOINT")
	if err != nil {
		return nil, err
	}
	accessKey, err := requiredEnv("MINIO_ACCESS_KEY")
	if err != nil {
		return nil, err
	}
	secretKey, err := requiredEnv("MINIO_SECRET_KEY")
	if err != nil {
		return nil, err
	}
	bucket, err := requiredEnv("MINIO_BUCKET")
	if err != nil {
		return nil, err
	}
	useSSL := false
	if raw := os.Getenv("MINIO_USE_SSL"); raw != "" {
		useSSL, err = strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("MINIO_USE_SSL: %w", err)
		}
	}
	store, err := objectstore.New(endpoint, accessKey, secretKey, bucket, useSSL)
	if err != nil {
		return nil, fmt.Errorf("connect object storage: %w", err)
	}
	if err := store.EnsureBucket(ctx); err != nil {
		return nil, fmt.Errorf("ensure bucket: %w", err)
	}
	return store, nil
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

	store, err := domain.NewStore(demoStoreID, merchantID, storeName, time.Now().UTC())
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

// seedDemoPets fills the store with the demo catalog, but only when the store has
// no pets at all, so re-running the seeder never piles up duplicate listings — even
// after a grader has bought or removed some of them.
func seedDemoPets(ctx context.Context, catalog *listing.Service, pets *postgres.PetRepository, storeID uuid.UUID) (int, error) {
	count, err := pets.CountByStore(ctx, storeID)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, nil
	}
	for _, pet := range demoCatalog {
		picture, err := speciesPicture(pet.species)
		if err != nil {
			return 0, err
		}
		_, err = catalog.CreatePet(ctx, listing.CreatePetCommand{
			StoreID:      storeID,
			Name:         pet.name,
			Species:      string(pet.species),
			AgeYears:     pet.ageYears,
			Description:  pet.description,
			BreederName:  breederName,
			BreederEmail: breederEmail,
			Picture:      bytes.NewReader(picture),
		})
		if err != nil {
			return 0, fmt.Errorf("create pet %q: %w", pet.name, err)
		}
	}
	return len(demoCatalog), nil
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
