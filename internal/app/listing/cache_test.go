package listing_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"roboticCrewChallenge/internal/adapter/postgres"
	"roboticCrewChallenge/internal/adapter/rediscache"
	"roboticCrewChallenge/internal/app/listing"
	"roboticCrewChallenge/internal/domain"
)

// countingPetRepo wraps the real repository and counts available-listing reads,
// so a test can prove a cache hit served a request without touching Postgres.
type countingPetRepo struct {
	domain.PetRepository
	listCalls atomic.Int64
}

func (c *countingPetRepo) ListAvailableByStore(ctx context.Context, storeID uuid.UUID, species *domain.Species, limit int, cursor string) ([]domain.Pet, string, error) {
	c.listCalls.Add(1)
	return c.PetRepository.ListAvailableByStore(ctx, storeID, species, limit, cursor)
}

func serviceWithSpy() (*listing.Service, *countingPetRepo) {
	spy := &countingPetRepo{PetRepository: postgres.NewPetRepository(harness.Pool, harness.Enc)}
	cache := rediscache.New(harness.RedisClient, harness.Enc, testLogger, testCacheTTL)
	return listing.NewService(spy, harness.PictureStore, cache), spy
}

// AC1: a repeated identical request is served from Redis and does not query Postgres.
func TestUnsoldPets_SecondReadServedFromCache(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	svc, spy := serviceWithSpy()
	storeID := seedStore(t)
	if _, err := svc.CreatePet(ctx, validCommand(storeID)); err != nil {
		t.Fatalf("create pet: %v", err)
	}

	if _, _, err := svc.UnsoldPets(ctx, storeID, 10, ""); err != nil {
		t.Fatalf("first read: %v", err)
	}
	if _, _, err := svc.UnsoldPets(ctx, storeID, 10, ""); err != nil {
		t.Fatalf("second read: %v", err)
	}

	if got := spy.listCalls.Load(); got != 1 {
		t.Fatalf("expected exactly one Postgres read (second served from cache), got %d", got)
	}
}

// AC2: creating a pet invalidates the cached listing so the change appears next read.
func TestCreatePet_InvalidatesCache(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	svc, spy := serviceWithSpy()
	storeID := seedStore(t)
	if _, err := svc.CreatePet(ctx, validCommand(storeID)); err != nil {
		t.Fatalf("create first pet: %v", err)
	}
	if _, _, err := svc.UnsoldPets(ctx, storeID, 10, ""); err != nil { // warms cache
		t.Fatalf("warm read: %v", err)
	}

	if _, err := svc.CreatePet(ctx, validCommand(storeID)); err != nil {
		t.Fatalf("create second pet: %v", err)
	}
	pets, _, err := svc.UnsoldPets(ctx, storeID, 10, "")
	if err != nil {
		t.Fatalf("post-create read: %v", err)
	}
	if len(pets) != 2 {
		t.Fatalf("new pet should appear after invalidation, got %d pets", len(pets))
	}
	if spy.listCalls.Load() != 2 {
		t.Fatalf("post-create read should miss the cache and re-query, got %d reads", spy.listCalls.Load())
	}
}

// AC2: removing a pet invalidates the cached listing.
func TestRemovePet_InvalidatesCache(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	svc, _ := serviceWithSpy()
	storeID := seedStore(t)
	first, err := svc.CreatePet(ctx, validCommand(storeID))
	if err != nil {
		t.Fatalf("create first pet: %v", err)
	}
	if _, err := svc.CreatePet(ctx, validCommand(storeID)); err != nil {
		t.Fatalf("create second pet: %v", err)
	}
	if _, _, err := svc.UnsoldPets(ctx, storeID, 10, ""); err != nil { // warms cache (2 pets)
		t.Fatalf("warm read: %v", err)
	}

	if _, err := svc.RemovePet(ctx, storeID, first.ID); err != nil {
		t.Fatalf("remove pet: %v", err)
	}
	pets, _, err := svc.UnsoldPets(ctx, storeID, 10, "")
	if err != nil {
		t.Fatalf("post-remove read: %v", err)
	}
	if len(pets) != 1 {
		t.Fatalf("removed pet should be gone after invalidation, got %d pets", len(pets))
	}
}

// AC2: a sale invalidates the cached listing. The customer purchase use case
// (a later issue) marks the pet sold and calls InvalidateStore; this stands in
// for that path with a direct repository purchase + cache invalidation.
func TestSale_InvalidatesCache(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	svc, _ := serviceWithSpy()
	repo := postgres.NewPetRepository(harness.Pool, harness.Enc)
	cache := rediscache.New(harness.RedisClient, harness.Enc, testLogger, testCacheTTL)
	storeID := seedStore(t)
	customerID := seedCustomer(t)

	pet, err := svc.CreatePet(ctx, validCommand(storeID))
	if err != nil {
		t.Fatalf("create pet: %v", err)
	}
	if _, _, err := svc.UnsoldPets(ctx, storeID, 10, ""); err != nil { // warms cache (1 available)
		t.Fatalf("warm read: %v", err)
	}

	if _, err := repo.Purchase(ctx, customerID, pet.ID); err != nil {
		t.Fatalf("purchase pet: %v", err)
	}
	cache.InvalidateStore(ctx, storeID)

	pets, _, err := svc.UnsoldPets(ctx, storeID, 10, "")
	if err != nil {
		t.Fatalf("post-sale read: %v", err)
	}
	if len(pets) != 0 {
		t.Fatalf("sold pet should be gone after invalidation, got %d pets", len(pets))
	}
}

// AC3: with Redis unavailable, reads still succeed from Postgres.
func TestUnsoldPets_SurvivesRedisDown(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	deadCache := rediscache.New(redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"}), harness.Enc, testLogger, testCacheTTL)
	svc := listing.NewService(postgres.NewPetRepository(harness.Pool, harness.Enc), harness.PictureStore, deadCache)
	storeID := seedStore(t)
	if _, err := svc.CreatePet(ctx, validCommand(storeID)); err != nil {
		t.Fatalf("create pet (write must survive a dead cache): %v", err)
	}

	pets, _, err := svc.UnsoldPets(ctx, storeID, 10, "")
	if err != nil {
		t.Fatalf("read must fall back to Postgres when Redis is down: %v", err)
	}
	if len(pets) != 1 {
		t.Fatalf("expected the Postgres-backed pet, got %d", len(pets))
	}
}
