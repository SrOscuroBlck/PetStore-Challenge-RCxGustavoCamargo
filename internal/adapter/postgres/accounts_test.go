package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"roboticCrewChallenge/internal/adapter/postgres/sqlcgen"
	"roboticCrewChallenge/internal/domain"
)

func seedCustomer(t *testing.T) uuid.UUID {
	t.Helper()
	id := uuid.New()
	customer, err := domain.NewCustomer(id, fmt.Sprintf("c-%s@example.com", id), "hashed", time.Now())
	if err != nil {
		t.Fatalf("build customer: %v", err)
	}
	if err := NewCustomerRepository(testPool, testEnc, testBlind).Create(context.Background(), customer); err != nil {
		t.Fatalf("create customer: %v", err)
	}
	return id
}

func TestMerchantRepository_CreateGetAndConflicts(t *testing.T) {
	ctx := context.Background()
	repo := NewMerchantRepository(testPool, testEnc, testBlind)
	id := uuid.New()
	email := fmt.Sprintf("merchant-%s@example.com", id)
	merchant, err := domain.NewMerchant(id, email, "hashed", time.Now())
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if err := repo.Create(ctx, merchant); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByEmail(ctx, email)
	if err != nil {
		t.Fatalf("get by email: %v", err)
	}
	if got.ID != id || got.Email != email {
		t.Fatalf("mismatch: %+v", got)
	}

	// duplicate email → typed conflict (AC3)
	dup, _ := domain.NewMerchant(uuid.New(), email, "hashed", time.Now())
	if err := repo.Create(ctx, dup); !errors.Is(err, domain.ErrEmailInUse) {
		t.Fatalf("want ErrEmailInUse, got %v", err)
	}
	// missing → typed not-found (AC3)
	if _, err := repo.GetByID(ctx, uuid.New()); !errors.Is(err, domain.ErrMerchantNotFound) {
		t.Fatalf("want ErrMerchantNotFound, got %v", err)
	}
	// email stored encrypted, not plaintext
	raw, err := sqlcgen.New(testPool).GetMerchantByID(ctx, id)
	if err != nil {
		t.Fatalf("raw: %v", err)
	}
	if string(raw.EmailEncrypted) == email {
		t.Fatal("merchant email is stored in plaintext")
	}
}

func TestCustomerRepository_CreateAndLookup(t *testing.T) {
	ctx := context.Background()
	repo := NewCustomerRepository(testPool, testEnc, testBlind)
	id := uuid.New()
	email := fmt.Sprintf("customer-%s@example.com", id)
	customer, _ := domain.NewCustomer(id, email, "hashed", time.Now())
	if err := repo.Create(ctx, customer); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := repo.GetByEmail(ctx, email)
	if err != nil {
		t.Fatalf("get by email: %v", err)
	}
	if got.ID != id {
		t.Fatalf("mismatch: %+v", got)
	}
	if _, err := repo.GetByID(ctx, uuid.New()); !errors.Is(err, domain.ErrCustomerNotFound) {
		t.Fatalf("want ErrCustomerNotFound, got %v", err)
	}
}

func TestStoreRepository_CreateAndConflict(t *testing.T) {
	ctx := context.Background()
	merchantRepo := NewMerchantRepository(testPool, testEnc, testBlind)
	storeRepo := NewStoreRepository(testPool)

	merchantID := uuid.New()
	merchant, _ := domain.NewMerchant(merchantID, fmt.Sprintf("ms-%s@example.com", merchantID), "hashed", time.Now())
	if err := merchantRepo.Create(ctx, merchant); err != nil {
		t.Fatalf("merchant: %v", err)
	}
	store, _ := domain.NewStore(uuid.New(), merchantID, "Store", time.Now())
	if err := storeRepo.Create(ctx, store); err != nil {
		t.Fatalf("store: %v", err)
	}
	got, err := storeRepo.GetByMerchantID(ctx, merchantID)
	if err != nil {
		t.Fatalf("get by merchant: %v", err)
	}
	if got.MerchantID != merchantID {
		t.Fatalf("mismatch: %+v", got)
	}
	// a second store for the same merchant → typed conflict
	second, _ := domain.NewStore(uuid.New(), merchantID, "Store2", time.Now())
	if err := storeRepo.Create(ctx, second); !errors.Is(err, domain.ErrStoreAlreadyExists) {
		t.Fatalf("want ErrStoreAlreadyExists, got %v", err)
	}
}
