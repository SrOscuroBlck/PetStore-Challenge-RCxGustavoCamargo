package testsupport

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"roboticCrewChallenge/internal/adapter/postgres"
	"roboticCrewChallenge/internal/domain"
)

// SeedStore creates a merchant and its store, returning the store id.
func (h *Harness) SeedStore(t *testing.T) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	merchantID := uuid.New()
	merchant, err := domain.NewMerchant(merchantID, fmt.Sprintf("m-%s@example.com", merchantID), "hash", time.Now().UTC())
	if err != nil {
		t.Fatalf("build merchant: %v", err)
	}
	if err := postgres.NewMerchantRepository(h.Pool, h.Enc, h.Blind).Create(ctx, merchant); err != nil {
		t.Fatalf("create merchant: %v", err)
	}
	storeID := uuid.New()
	store, err := domain.NewStore(storeID, merchantID, "Pluto's Pets", time.Now().UTC())
	if err != nil {
		t.Fatalf("build store: %v", err)
	}
	if err := postgres.NewStoreRepository(h.Pool).Create(ctx, store); err != nil {
		t.Fatalf("create store: %v", err)
	}
	return storeID
}

// SeedCustomer creates a customer and returns its id.
func (h *Harness) SeedCustomer(t *testing.T) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	customerID := uuid.New()
	customer, err := domain.NewCustomer(customerID, fmt.Sprintf("c-%s@example.com", customerID), "hash", time.Now().UTC())
	if err != nil {
		t.Fatalf("build customer: %v", err)
	}
	if err := postgres.NewCustomerRepository(h.Pool, h.Enc, h.Blind).Create(ctx, customer); err != nil {
		t.Fatalf("create customer: %v", err)
	}
	return customerID
}

// PNGBytes returns a minimal byte slice that sniffs as image/png.
func PNGBytes() []byte {
	return append([]byte("\x89PNG\r\n\x1a\n"), []byte("sniffable-png-pixel-data")...)
}
