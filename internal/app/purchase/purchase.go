package purchase

import (
	"context"

	"github.com/google/uuid"

	"roboticCrewChallenge/internal/domain"
)

type Service struct {
	pets  domain.PetRepository
	cache domain.CatalogCache
}

func NewService(pets domain.PetRepository, cache domain.CatalogCache) *Service {
	return &Service{pets: pets, cache: cache}
}

// PurchasePet buys a single available pet. The repository performs the atomic
// AVAILABLE->SOLD transition (idempotent for the same customer); on success the
// owning store's cached listing is invalidated so the pet leaves the catalog.
func (s *Service) PurchasePet(ctx context.Context, customerID, petID uuid.UUID) (domain.Pet, error) {
	pet, err := s.pets.Purchase(ctx, customerID, petID)
	if err != nil {
		return domain.Pet{}, err
	}
	s.cache.InvalidateStore(ctx, pet.StoreID)
	return pet, nil
}

// Checkout buys several pets atomically: the repository either purchases all of
// them or none (returning an error naming each unavailable pet). On success the
// cached listing of every distinct affected store is invalidated once.
func (s *Service) Checkout(ctx context.Context, customerID uuid.UUID, petIDs []uuid.UUID) ([]domain.Pet, error) {
	pets, err := s.pets.Checkout(ctx, customerID, petIDs)
	if err != nil {
		return nil, err
	}
	invalidated := make(map[uuid.UUID]struct{}, len(pets))
	for _, pet := range pets {
		if _, done := invalidated[pet.StoreID]; done {
			continue
		}
		invalidated[pet.StoreID] = struct{}{}
		s.cache.InvalidateStore(ctx, pet.StoreID)
	}
	return pets, nil
}
