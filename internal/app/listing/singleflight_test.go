package listing_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"roboticCrewChallenge/internal/app/listing"
	"roboticCrewChallenge/internal/domain"
)

// blockingRepo holds the first available-listing read open until released, so a
// test can pile concurrent callers behind it and observe whether they coalesce.
type blockingRepo struct {
	domain.PetRepository
	calls   atomic.Int64
	once    sync.Once
	started chan struct{}
	release chan struct{}
}

func (b *blockingRepo) ListAvailableByStore(context.Context, uuid.UUID, *domain.Species, int, string) ([]domain.Pet, string, error) {
	b.calls.Add(1)
	b.once.Do(func() { close(b.started) })
	<-b.release
	return nil, "", nil
}

type missCache struct{}

func (missCache) GetAvailable(context.Context, uuid.UUID, *domain.Species, int, string) (domain.CatalogPage, bool) {
	return domain.CatalogPage{}, false
}
func (missCache) SetAvailable(context.Context, uuid.UUID, *domain.Species, int, string, domain.CatalogPage) {
}
func (missCache) InvalidateStore(context.Context, uuid.UUID) {}

// A cache expiry under heavy concurrency must not stampede the database: many
// simultaneous misses for the same page collapse into a single read.
func TestAvailablePets_CoalescesConcurrentMisses(t *testing.T) {
	repo := &blockingRepo{started: make(chan struct{}), release: make(chan struct{})}
	svc := listing.NewService(repo, nil, missCache{})
	storeID := uuid.New()

	const callers = 50
	var wg sync.WaitGroup
	wg.Add(callers)
	for range callers {
		go func() {
			defer wg.Done()
			_, _, _ = svc.AvailablePets(context.Background(), storeID, nil, 10, "")
		}()
	}

	<-repo.started
	time.Sleep(50 * time.Millisecond)
	close(repo.release)
	wg.Wait()

	if got := repo.calls.Load(); got != 1 {
		t.Fatalf("expected %d concurrent misses to coalesce into 1 repository read, got %d", callers, got)
	}
}
