package listing

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"

	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/platform/id"
	"roboticCrewChallenge/internal/platform/pagination"
	"roboticCrewChallenge/internal/platform/picture"
)

type Service struct {
	pets     domain.PetRepository
	pictures domain.PictureStore
	cache    domain.CatalogCache
	newID    func() (uuid.UUID, error)
	now      func() time.Time
	loads    singleflight.Group
}

func NewService(pets domain.PetRepository, pictures domain.PictureStore, cache domain.CatalogCache) *Service {
	return &Service{
		pets:     pets,
		pictures: pictures,
		cache:    cache,
		newID:    id.New,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

type CreatePetCommand struct {
	StoreID      uuid.UUID
	Name         string
	Species      string
	AgeYears     int
	Description  string
	BreederName  string
	BreederEmail string
	Picture      io.Reader
}

// CreatePet validates and stores the picture before persisting the pet, so a
// rejected upload never leaves an object behind and the pet always references a
// real object key.
func (s *Service) CreatePet(ctx context.Context, cmd CreatePetCommand) (domain.Pet, error) {
	contentType, size, body, err := picture.Validate(cmd.Picture)
	if err != nil {
		return domain.Pet{}, err
	}
	objectKey, err := s.pictures.Upload(ctx, body, size, contentType)
	if err != nil {
		return domain.Pet{}, fmt.Errorf("upload picture: %w", err)
	}

	petID, err := s.newID()
	if err != nil {
		return domain.Pet{}, err
	}
	pet, err := domain.NewPet(domain.NewPetParams{
		ID:               petID,
		StoreID:          cmd.StoreID,
		Name:             cmd.Name,
		Species:          cmd.Species,
		AgeYears:         cmd.AgeYears,
		Description:      cmd.Description,
		BreederName:      cmd.BreederName,
		BreederEmail:     cmd.BreederEmail,
		PictureObjectKey: objectKey,
		CreatedAt:        s.now(),
	})
	if err != nil {
		return domain.Pet{}, err
	}
	if err := s.pets.Create(ctx, pet); err != nil {
		return domain.Pet{}, fmt.Errorf("create pet: %w", err)
	}
	s.cache.InvalidateStore(ctx, pet.StoreID)
	return pet, nil
}

func (s *Service) RemovePet(ctx context.Context, storeID, petID uuid.UUID) (domain.Pet, error) {
	pet, err := s.pets.Remove(ctx, storeID, petID)
	if err != nil {
		return domain.Pet{}, err
	}
	s.cache.InvalidateStore(ctx, storeID)
	return pet, nil
}

func (s *Service) SoldPets(ctx context.Context, storeID uuid.UUID, from, to time.Time, limit int, cursor string) ([]domain.Pet, string, error) {
	if err := validateCursor(cursor); err != nil {
		return nil, "", err
	}
	return s.pets.ListSoldByStore(ctx, storeID, from, to, limit, cursor)
}

// UnsoldPets lists a store's available pets (merchant view, unfiltered).
func (s *Service) UnsoldPets(ctx context.Context, storeID uuid.UUID, limit int, cursor string) ([]domain.Pet, string, error) {
	return s.availableByStore(ctx, storeID, nil, limit, cursor)
}

// AvailablePets lists a store's available pets (customer view), optionally
// filtered to a single species; a nil species returns every available pet.
func (s *Service) AvailablePets(ctx context.Context, storeID uuid.UUID, species *domain.Species, limit int, cursor string) ([]domain.Pet, string, error) {
	return s.availableByStore(ctx, storeID, species, limit, cursor)
}

func (s *Service) availableByStore(ctx context.Context, storeID uuid.UUID, species *domain.Species, limit int, cursor string) ([]domain.Pet, string, error) {
	if err := validateCursor(cursor); err != nil {
		return nil, "", err
	}
	if page, ok := s.cache.GetAvailable(ctx, storeID, species, limit, cursor); ok {
		return page.Pets, page.NextCursor, nil
	}
	page, err := s.loadAndCacheAvailable(ctx, storeID, species, limit, cursor)
	if err != nil {
		return nil, "", err
	}
	return page.Pets, page.NextCursor, nil
}

// loadAndCacheAvailable fetches a page from the source of truth and caches it,
// coalescing concurrent misses for the same page through singleflight so a cache
// expiry under heavy concurrency triggers one database read rather than a
// stampede of identical reads (see docs/PERFORMANCE.md).
func (s *Service) loadAndCacheAvailable(ctx context.Context, storeID uuid.UUID, species *domain.Species, limit int, cursor string) (domain.CatalogPage, error) {
	result, err, _ := s.loads.Do(availableLoadKey(storeID, species, limit, cursor), func() (any, error) {
		if page, ok := s.cache.GetAvailable(ctx, storeID, species, limit, cursor); ok {
			return page, nil
		}
		pets, nextCursor, err := s.pets.ListAvailableByStore(ctx, storeID, species, limit, cursor)
		if err != nil {
			return domain.CatalogPage{}, err
		}
		page := domain.CatalogPage{Pets: pets, NextCursor: nextCursor}
		s.cache.SetAvailable(ctx, storeID, species, limit, cursor, page)
		return page, nil
	})
	if err != nil {
		return domain.CatalogPage{}, err
	}
	return result.(domain.CatalogPage), nil
}

func availableLoadKey(storeID uuid.UUID, species *domain.Species, limit int, cursor string) string {
	speciesKey := "all"
	if species != nil {
		speciesKey = string(*species)
	}
	return fmt.Sprintf("%s|%s|%d|%s", storeID, speciesKey, limit, cursor)
}

func validateCursor(cursor string) error {
	if cursor == "" {
		return nil
	}
	if _, err := pagination.Decode(cursor); err != nil {
		return &domain.ValidationError{Field: "after", Msg: "is not a valid cursor"}
	}
	return nil
}
