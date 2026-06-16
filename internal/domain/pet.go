package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Pet struct {
	ID               uuid.UUID
	StoreID          uuid.UUID
	Name             string
	Species          Species
	AgeYears         int
	Description      string
	BreederName      string
	BreederEmail     string
	PictureObjectKey string
	Status           PetStatus
	CreatedAt        time.Time
	SoldAt           *time.Time
	SoldByCustomerID *uuid.UUID
	RemovedAt        *time.Time
}

type NewPetParams struct {
	ID               uuid.UUID
	StoreID          uuid.UUID
	Name             string
	Species          string
	AgeYears         int
	Description      string
	BreederName      string
	BreederEmail     string
	PictureObjectKey string
	CreatedAt        time.Time
}

func NewPet(p NewPetParams) (Pet, error) {
	if p.StoreID == uuid.Nil {
		return Pet{}, &ValidationError{Field: "storeID", Msg: "is required"}
	}
	if strings.TrimSpace(p.Name) == "" {
		return Pet{}, &ValidationError{Field: "name", Msg: "is required"}
	}
	species, err := ParseSpecies(p.Species)
	if err != nil {
		return Pet{}, err
	}
	if p.AgeYears < 0 {
		return Pet{}, &ValidationError{Field: "ageYears", Msg: "must not be negative"}
	}
	if strings.TrimSpace(p.Description) == "" {
		return Pet{}, &ValidationError{Field: "description", Msg: "is required"}
	}
	if strings.TrimSpace(p.BreederName) == "" {
		return Pet{}, &ValidationError{Field: "breederName", Msg: "is required"}
	}
	if err := validateEmail("breederEmail", p.BreederEmail); err != nil {
		return Pet{}, err
	}
	if strings.TrimSpace(p.PictureObjectKey) == "" {
		return Pet{}, &ValidationError{Field: "pictureObjectKey", Msg: "is required"}
	}

	return Pet{
		ID:               p.ID,
		StoreID:          p.StoreID,
		Name:             p.Name,
		Species:          species,
		AgeYears:         p.AgeYears,
		Description:      p.Description,
		BreederName:      p.BreederName,
		BreederEmail:     p.BreederEmail,
		PictureObjectKey: p.PictureObjectKey,
		Status:           PetStatusAvailable,
		CreatedAt:        p.CreatedAt,
	}, nil
}

func (p *Pet) MarkSold(customerID uuid.UUID, soldAt time.Time) error {
	if p.Status != PetStatusAvailable {
		return ErrPetUnavailable
	}
	p.Status = PetStatusSold
	p.SoldByCustomerID = &customerID
	p.SoldAt = &soldAt
	return nil
}

func (p *Pet) MarkRemoved(removedAt time.Time) error {
	if p.Status != PetStatusAvailable {
		return ErrPetNotRemovable
	}
	p.Status = PetStatusRemoved
	p.RemovedAt = &removedAt
	return nil
}
