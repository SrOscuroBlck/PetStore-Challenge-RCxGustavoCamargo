package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func validParams() NewPetParams {
	return NewPetParams{
		ID:               uuid.New(),
		StoreID:          uuid.New(),
		Name:             "Pluto",
		Species:          "DOG",
		AgeYears:         3,
		Description:      "Friendly and house-trained",
		BreederName:      "Jane Doe",
		BreederEmail:     "jane@example.com",
		PictureObjectKey: "pets/pluto.jpg",
		CreatedAt:        time.Now(),
	}
}

func newValidPet(t *testing.T) Pet {
	t.Helper()
	pet, err := NewPet(validParams())
	if err != nil {
		t.Fatalf("expected a valid pet, got %v", err)
	}
	return pet
}

func TestNewPet_ValidStartsAvailable(t *testing.T) {
	pet := newValidPet(t)
	if pet.Status != PetStatusAvailable {
		t.Fatalf("expected status AVAILABLE, got %s", pet.Status)
	}
}

func TestNewPet_RejectsUnknownSpecies(t *testing.T) {
	p := validParams()
	p.Species = "TIGER"
	_, err := NewPet(p)
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "species" {
		t.Fatalf("expected a species ValidationError, got %v", err)
	}
}

func TestNewPet_RejectsEmptyName(t *testing.T) {
	p := validParams()
	p.Name = "   "
	_, err := NewPet(p)
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "name" {
		t.Fatalf("expected a name ValidationError, got %v", err)
	}
}

func TestNewPet_RejectsNegativeAge(t *testing.T) {
	p := validParams()
	p.AgeYears = -1
	_, err := NewPet(p)
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "ageYears" {
		t.Fatalf("expected an ageYears ValidationError, got %v", err)
	}
}

func TestNewPet_RejectsInvalidBreederEmail(t *testing.T) {
	p := validParams()
	p.BreederEmail = "not-an-email"
	_, err := NewPet(p)
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "breederEmail" {
		t.Fatalf("expected a breederEmail ValidationError, got %v", err)
	}
}

func TestPet_MarkSold_OnlyFromAvailable(t *testing.T) {
	pet := newValidPet(t)
	if err := pet.MarkSold(uuid.New(), time.Now()); err != nil {
		t.Fatalf("first sale should succeed, got %v", err)
	}
	if pet.Status != PetStatusSold || pet.SoldByCustomerID == nil || pet.SoldAt == nil {
		t.Fatal("expected the pet to record the sale")
	}
	if err := pet.MarkSold(uuid.New(), time.Now()); !errors.Is(err, ErrPetUnavailable) {
		t.Fatalf("second sale should fail with ErrPetUnavailable, got %v", err)
	}
}

func TestPet_MarkRemoved_OnlyFromAvailable(t *testing.T) {
	pet := newValidPet(t)
	if err := pet.MarkSold(uuid.New(), time.Now()); err != nil {
		t.Fatalf("sale should succeed, got %v", err)
	}
	if err := pet.MarkRemoved(time.Now()); !errors.Is(err, ErrPetNotRemovable) {
		t.Fatalf("removing a sold pet should fail with ErrPetNotRemovable, got %v", err)
	}
}

func TestParseSpecies(t *testing.T) {
	if _, err := ParseSpecies("CAT"); err != nil {
		t.Fatalf("CAT should parse, got %v", err)
	}
	if _, err := ParseSpecies("tiger"); err == nil {
		t.Fatal("tiger should not parse")
	}
}
