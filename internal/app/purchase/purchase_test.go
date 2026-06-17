package purchase_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"

	"roboticCrewChallenge/internal/adapter/postgres"
	"roboticCrewChallenge/internal/app/purchase"
	"roboticCrewChallenge/internal/domain"
)

func newService(cache domain.CatalogCache) *purchase.Service {
	return purchase.NewService(postgres.NewPetRepository(harness.Pool, harness.Enc), cache)
}

func TestPurchasePet_SucceedsAndInvalidatesStore(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	cache := &spyCache{}
	svc := newService(cache)
	storeID := seedStore(t)
	pet := seedPet(t, storeID)
	customerID := seedCustomer(t)

	sold, err := svc.PurchasePet(ctx, customerID, pet.ID)
	if err != nil {
		t.Fatalf("purchase: %v", err)
	}
	if sold.Status != domain.PetStatusSold || sold.SoldByCustomerID == nil || *sold.SoldByCustomerID != customerID {
		t.Fatalf("unexpected sold pet: %+v", sold)
	}
	if calls := cache.calls(); len(calls) != 1 || calls[0] != storeID {
		t.Fatalf("expected one invalidation of store %s, got %v", storeID, calls)
	}
}

func TestPurchasePet_UnavailableIsRejectedAndNotInvalidated(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	cache := &spyCache{}
	svc := newService(cache)
	pet := seedPet(t, seedStore(t))

	if _, err := svc.PurchasePet(ctx, seedCustomer(t), pet.ID); err != nil {
		t.Fatalf("first purchase: %v", err)
	}
	if _, err := svc.PurchasePet(ctx, seedCustomer(t), pet.ID); !errors.Is(err, domain.ErrPetUnavailable) {
		t.Fatalf("second customer should get ErrPetUnavailable, got %v", err)
	}
	// Only the first (successful) purchase invalidated; the failed one did not.
	if calls := cache.calls(); len(calls) != 1 {
		t.Fatalf("a failed purchase must not invalidate the cache, got %d calls", len(calls))
	}
}

func TestPurchasePet_IdempotentForSameCustomer(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	svc := newService(&spyCache{})
	pet := seedPet(t, seedStore(t))
	customerID := seedCustomer(t)

	if _, err := svc.PurchasePet(ctx, customerID, pet.ID); err != nil {
		t.Fatalf("first purchase: %v", err)
	}
	if _, err := svc.PurchasePet(ctx, customerID, pet.ID); err != nil {
		t.Fatalf("idempotent re-purchase by the same customer should succeed, got %v", err)
	}
}

func TestCheckout_AtomicSuccessInvalidatesEachStoreOnce(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	cache := &spyCache{}
	svc := newService(cache)
	storeA := seedStore(t)
	storeB := seedStore(t)
	petA := seedPet(t, storeA)
	petB := seedPet(t, storeB)
	customerID := seedCustomer(t)

	pets, err := svc.Checkout(ctx, customerID, []uuid.UUID{petA.ID, petB.ID})
	if err != nil {
		t.Fatalf("checkout: %v", err)
	}
	if len(pets) != 2 {
		t.Fatalf("expected 2 purchased pets, got %d", len(pets))
	}
	calls := cache.calls()
	if len(calls) != 2 {
		t.Fatalf("expected one invalidation per distinct store, got %v", calls)
	}
	seen := map[uuid.UUID]bool{calls[0]: true, calls[1]: true}
	if !seen[storeA] || !seen[storeB] {
		t.Fatalf("both stores must be invalidated, got %v", calls)
	}
}

func TestCheckout_UnavailablePetRollsBackAndNamesIt(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	cache := &spyCache{}
	svc := newService(cache)
	repo := postgres.NewPetRepository(harness.Pool, harness.Enc)
	storeID := seedStore(t)
	available := seedPet(t, storeID)
	taken := seedPet(t, storeID)
	if _, err := repo.Purchase(ctx, seedCustomer(t), taken.ID); err != nil {
		t.Fatalf("pre-sell: %v", err)
	}

	_, err := svc.Checkout(ctx, seedCustomer(t), []uuid.UUID{available.ID, taken.ID})
	var unavailable *domain.UnavailablePetsError
	if !errors.As(err, &unavailable) {
		t.Fatalf("expected *UnavailablePetsError, got %v", err)
	}
	if len(unavailable.Pets) != 1 || unavailable.Pets[0].ID != taken.ID || unavailable.Pets[0].Name == "" {
		t.Fatalf("error must name the unavailable pet, got %+v", unavailable.Pets)
	}
	// All-or-nothing: the available pet stays AVAILABLE and nothing was invalidated.
	got, err := repo.GetByID(ctx, storeID, available.ID)
	if err != nil {
		t.Fatalf("get available pet: %v", err)
	}
	if got.Status != domain.PetStatusAvailable {
		t.Fatalf("available pet must remain AVAILABLE after rollback, got %s", got.Status)
	}
	if calls := cache.calls(); len(calls) != 0 {
		t.Fatalf("a failed checkout must not invalidate the cache, got %v", calls)
	}
}

// AC4: under concurrent purchase of the same pet, exactly one customer wins.
func TestPurchasePet_OnlyOneWinnerUnderConcurrency(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	svc := newService(&spyCache{})
	storeID := seedStore(t)
	pet := seedPet(t, storeID)

	const n = 12
	customers := make([]uuid.UUID, n)
	for i := range customers {
		customers[i] = seedCustomer(t)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	start := make(chan struct{})
	wins, losses := 0, 0
	for _, customerID := range customers {
		wg.Add(1)
		go func(customerID uuid.UUID) {
			defer wg.Done()
			<-start
			_, err := svc.PurchasePet(ctx, customerID, pet.ID)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				wins++
			case errors.Is(err, domain.ErrPetUnavailable):
				losses++
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}(customerID)
	}
	close(start)
	wg.Wait()

	if wins != 1 {
		t.Fatalf("expected exactly one winner, got %d (losses=%d)", wins, losses)
	}
}
