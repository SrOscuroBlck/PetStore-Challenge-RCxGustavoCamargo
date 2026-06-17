package rediscache

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/platform/crypto"
)

const breederEmail = "jane@example.com"

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func samplePage(t *testing.T, storeID uuid.UUID) domain.CatalogPage {
	t.Helper()
	pet, err := domain.NewPet(domain.NewPetParams{
		ID:               uuid.New(),
		StoreID:          storeID,
		Name:             "Pluto",
		Species:          "DOG",
		AgeYears:         3,
		Description:      "Friendly",
		BreederName:      "Jane Doe",
		BreederEmail:     breederEmail,
		PictureObjectKey: "pets/pluto",
		CreatedAt:        time.Now().UTC().Truncate(time.Microsecond),
	})
	if err != nil {
		t.Fatalf("build pet: %v", err)
	}
	return domain.CatalogPage{Pets: []domain.Pet{pet}, NextCursor: "next"}
}

func newCache() *Cache {
	return New(harness.RedisClient, harness.Enc, discardLogger, time.Minute)
}

func TestCache_PayloadIsEncryptedAtRest(t *testing.T) {
	requireRedis(t)
	ctx := context.Background()
	cache := newCache()
	storeID := uuid.New()
	page := samplePage(t, storeID)

	cache.SetAvailable(ctx, storeID, 10, "", page)

	raw, err := harness.RedisClient.Get(ctx, pageKey(storeID, 0, 10, "")).Bytes()
	if err != nil {
		t.Fatalf("read raw cache value: %v", err)
	}
	if bytes.Contains(raw, []byte(breederEmail)) || bytes.Contains(raw, []byte("Jane Doe")) {
		t.Fatal("breeder PII must not appear in plaintext in the cached value")
	}

	got, ok := cache.GetAvailable(ctx, storeID, 10, "")
	if !ok {
		t.Fatal("expected a cache hit")
	}
	if len(got.Pets) != 1 || got.Pets[0].BreederEmail != breederEmail || got.NextCursor != "next" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestCache_InvalidateStoreOrphansOldPages(t *testing.T) {
	requireRedis(t)
	ctx := context.Background()
	cache := newCache()
	storeID := uuid.New()

	cache.SetAvailable(ctx, storeID, 10, "", samplePage(t, storeID))
	if _, ok := cache.GetAvailable(ctx, storeID, 10, ""); !ok {
		t.Fatal("expected a hit before invalidation")
	}

	cache.InvalidateStore(ctx, storeID)

	if _, ok := cache.GetAvailable(ctx, storeID, 10, ""); ok {
		t.Fatal("expected a miss after invalidation (generation advanced)")
	}
}

func TestCache_ColdStoreMisses(t *testing.T) {
	requireRedis(t)
	cache := newCache()
	if _, ok := cache.GetAvailable(context.Background(), uuid.New(), 10, ""); ok {
		t.Fatal("a cold store must miss")
	}
}

func TestCache_NonFatalWhenRedisUnreachable(t *testing.T) {
	ctx := context.Background()
	key := make([]byte, crypto.KeySize)
	for i := range key {
		key[i] = byte(i + 1)
	}
	enc, err := crypto.NewEncryptor(key)
	if err != nil {
		t.Fatalf("encryptor: %v", err)
	}
	dead := New(redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"}), enc, discardLogger, time.Minute)
	storeID := uuid.New()

	if _, ok := dead.GetAvailable(ctx, storeID, 10, ""); ok {
		t.Fatal("a dead client must report a miss, not a hit")
	}
	// Writes and invalidation must not panic against a dead client.
	dead.SetAvailable(ctx, storeID, 10, "", samplePage(t, storeID))
	dead.InvalidateStore(ctx, storeID)
}
