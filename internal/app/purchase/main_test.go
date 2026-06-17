package purchase_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"roboticCrewChallenge/internal/adapter/postgres"
	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/testsupport"
)

var harness *testsupport.Harness

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()
	schema, err := filepath.Abs(filepath.Join("..", "..", "..", "db", "schema", "schema.sql"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "resolve schema path:", err)
		return 1
	}
	h, cleanup, ok, err := testsupport.Start(ctx, testsupport.Options{SchemaPath: schema, WithPostgres: true})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if !ok {
		return m.Run()
	}
	defer cleanup()
	harness = h
	return m.Run()
}

func requireInfra(t *testing.T) {
	t.Helper()
	if harness == nil {
		t.Skip("postgres container unavailable")
	}
}

func seedStore(t *testing.T) uuid.UUID    { return harness.SeedStore(t) }
func seedCustomer(t *testing.T) uuid.UUID { return harness.SeedCustomer(t) }

func seedPet(t *testing.T, storeID uuid.UUID) domain.Pet {
	t.Helper()
	pet, err := domain.NewPet(domain.NewPetParams{
		ID:               uuid.New(),
		StoreID:          storeID,
		Name:             "Pluto",
		Species:          "DOG",
		AgeYears:         3,
		Description:      "Friendly",
		BreederName:      "Jane Doe",
		BreederEmail:     "jane@example.com",
		PictureObjectKey: "pets/pluto",
		CreatedAt:        time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("build pet: %v", err)
	}
	if err := postgres.NewPetRepository(harness.Pool, harness.Enc).Create(context.Background(), pet); err != nil {
		t.Fatalf("create pet: %v", err)
	}
	return pet
}

// spyCache records which stores were invalidated so a test can assert the
// service invalidated exactly the affected stores. Reads always miss.
type spyCache struct {
	mu          sync.Mutex
	invalidated []uuid.UUID
}

func (s *spyCache) GetAvailable(context.Context, uuid.UUID, *domain.Species, int, string) (domain.CatalogPage, bool) {
	return domain.CatalogPage{}, false
}

func (s *spyCache) SetAvailable(context.Context, uuid.UUID, *domain.Species, int, string, domain.CatalogPage) {
}

func (s *spyCache) InvalidateStore(_ context.Context, storeID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invalidated = append(s.invalidated, storeID)
}

func (s *spyCache) calls() []uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]uuid.UUID(nil), s.invalidated...)
}
