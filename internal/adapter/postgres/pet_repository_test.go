package postgres

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"roboticCrewChallenge/internal/adapter/postgres/sqlcgen"
	"roboticCrewChallenge/internal/domain"
)

func TestPetRepository_Checkout_DeduplicatesIDs(t *testing.T) {
	ctx := context.Background()
	repo := NewPetRepository(testPool, testEnc)
	storeID := seedStore(t)
	pet := newPet(t, storeID)
	mustCreate(t, repo, pet)

	// Same pet listed twice in the cart must resolve to a single purchase.
	purchased, err := repo.Checkout(ctx, seedCustomer(t), []uuid.UUID{pet.ID, pet.ID})
	if err != nil {
		t.Fatalf("checkout with duplicate ids should succeed, got %v", err)
	}
	if len(purchased) != 1 {
		t.Fatalf("want 1 purchased pet, got %d", len(purchased))
	}
	got, err := repo.GetByID(ctx, storeID, pet.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != domain.PetStatusSold {
		t.Fatalf("want SOLD, got %s", got.Status)
	}
}

func TestPetRepository_ListSold_ByRange(t *testing.T) {
	ctx := context.Background()
	repo := NewPetRepository(testPool, testEnc)
	storeID := seedStore(t)
	const total = 4
	for i := 0; i < total; i++ {
		p := newPet(t, storeID)
		mustCreate(t, repo, p)
		if _, err := repo.Purchase(ctx, seedCustomer(t), p.ID); err != nil {
			t.Fatalf("purchase: %v", err)
		}
	}

	from := time.Now().UTC().Add(-time.Hour)
	to := time.Now().UTC().Add(time.Hour)
	seen := map[uuid.UUID]bool{}
	cursor := ""
	for pages := 0; pages < total+1; pages++ {
		pets, next, err := repo.ListSoldByStore(ctx, storeID, from, to, 2, cursor)
		if err != nil {
			t.Fatalf("list sold: %v", err)
		}
		for _, p := range pets {
			if p.Status != domain.PetStatusSold {
				t.Fatalf("expected SOLD, got %s", p.Status)
			}
			if seen[p.ID] {
				t.Fatalf("duplicate pet %s across pages", p.ID)
			}
			seen[p.ID] = true
		}
		if next == "" {
			break
		}
		cursor = next
	}
	if len(seen) != total {
		t.Fatalf("want %d sold pets, got %d", total, len(seen))
	}

	// a window in the future excludes them all
	future := time.Now().UTC().Add(24 * time.Hour)
	pets, _, err := repo.ListSoldByStore(ctx, storeID, future, future.Add(time.Hour), 10, "")
	if err != nil {
		t.Fatalf("list sold (future): %v", err)
	}
	if len(pets) != 0 {
		t.Fatalf("want 0 sold pets in future window, got %d", len(pets))
	}
}

func mustCreate(t *testing.T, repo *PetRepository, pet domain.Pet) {
	t.Helper()
	if err := repo.Create(context.Background(), pet); err != nil {
		t.Fatalf("create pet: %v", err)
	}
}

func TestPetRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	repo := NewPetRepository(testPool, testEnc)
	storeID := seedStore(t)
	pet := newPet(t, storeID)
	mustCreate(t, repo, pet)

	got, err := repo.GetByID(ctx, storeID, pet.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != pet.Name || got.BreederEmail != pet.BreederEmail || got.Status != domain.PetStatusAvailable {
		t.Fatalf("round-trip mismatch: %+v", got)
	}

	// PII must be stored encrypted, not as plaintext.
	raw, err := sqlcgen.New(testPool).GetPetByIDUnscoped(ctx, pet.ID)
	if err != nil {
		t.Fatalf("raw get: %v", err)
	}
	if string(raw.BreederEmailEncrypted) == pet.BreederEmail {
		t.Fatal("breeder email is stored in plaintext")
	}
	if string(raw.BreederNameEncrypted) == pet.BreederName {
		t.Fatal("breeder name is stored in plaintext")
	}
}

func TestPetRepository_GetByID_NotFound(t *testing.T) {
	_, err := NewPetRepository(testPool, testEnc).GetByID(context.Background(), seedStore(t), uuid.New())
	if !errors.Is(err, domain.ErrPetNotFound) {
		t.Fatalf("want ErrPetNotFound, got %v", err)
	}
}

func TestPetRepository_Remove(t *testing.T) {
	ctx := context.Background()
	repo := NewPetRepository(testPool, testEnc)
	storeID := seedStore(t)
	pet := newPet(t, storeID)
	mustCreate(t, repo, pet)

	removed, err := repo.Remove(ctx, storeID, pet.ID)
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if removed.Status != domain.PetStatusRemoved {
		t.Fatalf("want REMOVED, got %s", removed.Status)
	}
	if _, err := repo.Remove(ctx, storeID, pet.ID); !errors.Is(err, domain.ErrPetNotRemovable) {
		t.Fatalf("re-remove want ErrPetNotRemovable, got %v", err)
	}
	if _, err := repo.Remove(ctx, storeID, uuid.New()); !errors.Is(err, domain.ErrPetNotFound) {
		t.Fatalf("remove missing want ErrPetNotFound, got %v", err)
	}
}

func TestPetRepository_Purchase_Idempotent(t *testing.T) {
	ctx := context.Background()
	repo := NewPetRepository(testPool, testEnc)
	pet := newPet(t, seedStore(t))
	mustCreate(t, repo, pet)
	customerID := seedCustomer(t)

	sold, err := repo.Purchase(ctx, customerID, pet.ID)
	if err != nil {
		t.Fatalf("purchase: %v", err)
	}
	if sold.Status != domain.PetStatusSold || sold.SoldByCustomerID == nil || *sold.SoldByCustomerID != customerID {
		t.Fatalf("unexpected sold pet: %+v", sold)
	}
	if _, err := repo.Purchase(ctx, customerID, pet.ID); err != nil {
		t.Fatalf("idempotent re-purchase should succeed, got %v", err)
	}
	if _, err := repo.Purchase(ctx, seedCustomer(t), pet.ID); !errors.Is(err, domain.ErrPetUnavailable) {
		t.Fatalf("other customer want ErrPetUnavailable, got %v", err)
	}
}

func TestPetRepository_Purchase_OnlyOneWinnerUnderConcurrency(t *testing.T) {
	ctx := context.Background()
	repo := NewPetRepository(testPool, testEnc)
	storeID := seedStore(t)
	pet := newPet(t, storeID)
	mustCreate(t, repo, pet)

	const n = 12
	customers := make([]uuid.UUID, n)
	for i := range customers {
		customers[i] = seedCustomer(t)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	start := make(chan struct{})
	wins, losses := 0, 0
	for _, cust := range customers {
		wg.Add(1)
		go func(cust uuid.UUID) {
			defer wg.Done()
			<-start
			_, err := repo.Purchase(ctx, cust, pet.ID)
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
		}(cust)
	}
	close(start)
	wg.Wait()

	if wins != 1 {
		t.Fatalf("expected exactly one winner, got %d (losses=%d)", wins, losses)
	}
	final, err := repo.GetByID(ctx, storeID, pet.ID)
	if err != nil {
		t.Fatalf("final get: %v", err)
	}
	if final.Status != domain.PetStatusSold {
		t.Fatalf("final status want SOLD, got %s", final.Status)
	}
}

func TestPetRepository_Checkout_AllOrNothing(t *testing.T) {
	ctx := context.Background()
	repo := NewPetRepository(testPool, testEnc)
	storeID := seedStore(t)
	a1 := newPet(t, storeID)
	a2 := newPet(t, storeID)
	taken := newPet(t, storeID)
	mustCreate(t, repo, a1)
	mustCreate(t, repo, a2)
	mustCreate(t, repo, taken)
	if _, err := repo.Purchase(ctx, seedCustomer(t), taken.ID); err != nil {
		t.Fatalf("pre-sell: %v", err)
	}

	customerID := seedCustomer(t)
	_, err := repo.Checkout(ctx, customerID, []uuid.UUID{a1.ID, taken.ID, a2.ID})
	var unavail *domain.UnavailablePetsError
	if !errors.As(err, &unavail) {
		t.Fatalf("want UnavailablePetsError, got %v", err)
	}
	if len(unavail.Pets) != 1 || unavail.Pets[0].ID != taken.ID || unavail.Pets[0].Name == "" {
		t.Fatalf("want taken named unavailable, got %+v", unavail.Pets)
	}
	// rollback: the available pets must remain available
	for _, id := range []uuid.UUID{a1.ID, a2.ID} {
		p, err := repo.GetByID(ctx, storeID, id)
		if err != nil {
			t.Fatalf("get %s: %v", id, err)
		}
		if p.Status != domain.PetStatusAvailable {
			t.Fatalf("pet %s should remain AVAILABLE after rollback, got %s", id, p.Status)
		}
	}

	purchased, err := repo.Checkout(ctx, customerID, []uuid.UUID{a1.ID, a2.ID})
	if err != nil {
		t.Fatalf("checkout available: %v", err)
	}
	if len(purchased) != 2 {
		t.Fatalf("want 2 purchased, got %d", len(purchased))
	}
}

func TestPetRepository_ListAvailable_Pagination(t *testing.T) {
	ctx := context.Background()
	repo := NewPetRepository(testPool, testEnc)
	storeID := seedStore(t)
	const total = 5
	for i := 0; i < total; i++ {
		p := newPet(t, storeID)
		p.CreatedAt = time.Now().UTC().Add(time.Duration(i) * time.Millisecond)
		mustCreate(t, repo, p)
	}

	seen := map[uuid.UUID]bool{}
	cursor := ""
	for pages := 0; pages < total+1; pages++ {
		pets, next, err := repo.ListAvailableByStore(ctx, storeID, nil, 2, cursor)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		for _, p := range pets {
			if seen[p.ID] {
				t.Fatalf("duplicate pet %s across pages", p.ID)
			}
			seen[p.ID] = true
		}
		if next == "" {
			break
		}
		cursor = next
	}
	if len(seen) != total {
		t.Fatalf("want %d pets across pages, got %d", total, len(seen))
	}
}
