package listing_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"roboticCrewChallenge/internal/adapter/postgres"
	"roboticCrewChallenge/internal/app/listing"
	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/platform/picture"
)

func validCommand(storeID uuid.UUID) listing.CreatePetCommand {
	return listing.CreatePetCommand{
		StoreID:      storeID,
		Name:         "Pluto",
		Species:      "DOG",
		AgeYears:     3,
		Description:  "Friendly",
		BreederName:  "Jane Doe",
		BreederEmail: "jane@example.com",
		Picture:      bytes.NewReader(pngBytes),
	}
}

func TestCreatePet_StoresAvailablePetWithKey(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	svc := newService()
	storeID := seedStore(t)

	pet, err := svc.CreatePet(ctx, validCommand(storeID))
	if err != nil {
		t.Fatalf("create pet: %v", err)
	}
	if pet.Status != domain.PetStatusAvailable {
		t.Fatalf("status = %s, want AVAILABLE", pet.Status)
	}
	if pet.CreatedAt.IsZero() {
		t.Fatal("createdAt must be set")
	}
	if pet.PictureObjectKey == "" {
		t.Fatal("picture object key must be set")
	}
	content, err := harness.PictureStore.Get(ctx, pet.PictureObjectKey)
	if err != nil {
		t.Fatalf("uploaded object should be retrievable: %v", err)
	}
	_ = content.Body.Close()
}

func TestCreatePet_RejectsBadInput(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	svc := newService()
	storeID := seedStore(t)

	t.Run("oversize picture", func(t *testing.T) {
		cmd := validCommand(storeID)
		oversize := make([]byte, picture.MaxPictureBytes+1)
		copy(oversize, pngBytes)
		cmd.Picture = bytes.NewReader(oversize)
		if _, err := svc.CreatePet(ctx, cmd); !errors.Is(err, domain.ErrPictureTooLarge) {
			t.Fatalf("expected ErrPictureTooLarge, got %v", err)
		}
	})

	t.Run("non-image picture", func(t *testing.T) {
		cmd := validCommand(storeID)
		cmd.Picture = bytes.NewReader([]byte("this is plainly not an image"))
		if _, err := svc.CreatePet(ctx, cmd); !errors.Is(err, domain.ErrUnsupportedPictureType) {
			t.Fatalf("expected ErrUnsupportedPictureType, got %v", err)
		}
	})

	t.Run("invalid species", func(t *testing.T) {
		cmd := validCommand(storeID)
		cmd.Species = "DRAGON"
		var ve *domain.ValidationError
		if _, err := svc.CreatePet(ctx, cmd); !errors.As(err, &ve) {
			t.Fatalf("expected *ValidationError, got %v", err)
		}
	})
}

func TestRemovePet_AvailableBecomesRemoved(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	svc := newService()
	storeID := seedStore(t)

	created, err := svc.CreatePet(ctx, validCommand(storeID))
	if err != nil {
		t.Fatalf("create pet: %v", err)
	}
	removed, err := svc.RemovePet(ctx, storeID, created.ID)
	if err != nil {
		t.Fatalf("remove pet: %v", err)
	}
	if removed.Status != domain.PetStatusRemoved {
		t.Fatalf("status = %s, want REMOVED", removed.Status)
	}
}

func TestRemovePet_AlreadySoldIsRejected(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	svc := newService()
	repo := postgres.NewPetRepository(harness.Pool, harness.Enc)
	storeID := seedStore(t)
	customerID := seedCustomer(t)

	created, err := svc.CreatePet(ctx, validCommand(storeID))
	if err != nil {
		t.Fatalf("create pet: %v", err)
	}
	if _, err := repo.Purchase(ctx, customerID, created.ID); err != nil {
		t.Fatalf("purchase pet: %v", err)
	}
	if _, err := svc.RemovePet(ctx, storeID, created.ID); !errors.Is(err, domain.ErrPetNotRemovable) {
		t.Fatalf("expected ErrPetNotRemovable, got %v", err)
	}
}

func TestRemovePet_CrossStoreIsNotFound(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	svc := newService()
	storeA := seedStore(t)
	storeB := seedStore(t)

	pet, err := svc.CreatePet(ctx, validCommand(storeA))
	if err != nil {
		t.Fatalf("create pet: %v", err)
	}
	if _, err := svc.RemovePet(ctx, storeB, pet.ID); !errors.Is(err, domain.ErrPetNotFound) {
		t.Fatalf("expected ErrPetNotFound for a cross-store removal, got %v", err)
	}
}

func TestSoldPets_InclusiveRange(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	svc := newService()
	repo := postgres.NewPetRepository(harness.Pool, harness.Enc)
	storeID := seedStore(t)
	customerID := seedCustomer(t)

	created, err := svc.CreatePet(ctx, validCommand(storeID))
	if err != nil {
		t.Fatalf("create pet: %v", err)
	}
	sold, err := repo.Purchase(ctx, customerID, created.ID)
	if err != nil {
		t.Fatalf("purchase pet: %v", err)
	}
	soldAt := *sold.SoldAt

	in, _, err := svc.SoldPets(ctx, storeID, soldAt, soldAt, 10, "")
	if err != nil {
		t.Fatalf("sold pets: %v", err)
	}
	if len(in) != 1 || in[0].ID != created.ID {
		t.Fatalf("inclusive [soldAt,soldAt] range should return the pet, got %d", len(in))
	}

	out, _, err := svc.SoldPets(ctx, storeID, soldAt.Add(time.Millisecond), soldAt.Add(time.Hour), 10, "")
	if err != nil {
		t.Fatalf("sold pets: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("range starting after soldAt should exclude the pet, got %d", len(out))
	}
}

func TestUnsoldPets_OnlyAvailablePaginated(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	svc := newService()
	repo := postgres.NewPetRepository(harness.Pool, harness.Enc)
	storeID := seedStore(t)
	customerID := seedCustomer(t)

	for range 3 {
		if _, err := svc.CreatePet(ctx, validCommand(storeID)); err != nil {
			t.Fatalf("create pet: %v", err)
		}
	}
	soldPet, err := svc.CreatePet(ctx, validCommand(storeID))
	if err != nil {
		t.Fatalf("create pet: %v", err)
	}
	if _, err := repo.Purchase(ctx, customerID, soldPet.ID); err != nil {
		t.Fatalf("purchase pet: %v", err)
	}

	firstPage, next, err := svc.UnsoldPets(ctx, storeID, 2, "")
	if err != nil {
		t.Fatalf("unsold pets: %v", err)
	}
	if len(firstPage) != 2 || next == "" {
		t.Fatalf("first page should be full with a next cursor, got %d pets next=%q", len(firstPage), next)
	}
	for _, pet := range firstPage {
		if pet.Status != domain.PetStatusAvailable {
			t.Fatalf("unsold list must contain only AVAILABLE pets, got %s", pet.Status)
		}
	}

	secondPage, next2, err := svc.UnsoldPets(ctx, storeID, 2, next)
	if err != nil {
		t.Fatalf("unsold pets page 2: %v", err)
	}
	if len(secondPage) != 1 || next2 != "" {
		t.Fatalf("second page should hold the last available pet and end, got %d pets next=%q", len(secondPage), next2)
	}
}

func TestUnsoldPets_RejectsBadCursor(t *testing.T) {
	requireInfra(t)
	ctx := context.Background()
	svc := newService()
	storeID := seedStore(t)

	var ve *domain.ValidationError
	if _, _, err := svc.UnsoldPets(ctx, storeID, 10, "not-a-cursor"); !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for a bad cursor, got %v", err)
	}
}
