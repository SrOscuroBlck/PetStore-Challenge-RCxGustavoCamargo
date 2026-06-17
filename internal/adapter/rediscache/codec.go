package rediscache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"roboticCrewChallenge/internal/domain"
)

// petDTO is the adapter-owned wire shape for a cached pet. Keeping it separate
// from domain.Pet means the cache's serialization format is decoupled from the
// domain model: the domain struct stays free of storage concerns and a field
// change there cannot silently alter what is persisted in Redis.
type petDTO struct {
	ID               string     `json:"id"`
	StoreID          string     `json:"storeId"`
	Name             string     `json:"name"`
	Species          string     `json:"species"`
	AgeYears         int        `json:"ageYears"`
	Description      string     `json:"description"`
	BreederName      string     `json:"breederName"`
	BreederEmail     string     `json:"breederEmail"`
	PictureObjectKey string     `json:"pictureObjectKey"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"createdAt"`
	SoldAt           *time.Time `json:"soldAt,omitempty"`
	SoldByCustomerID *string    `json:"soldByCustomerId,omitempty"`
	RemovedAt        *time.Time `json:"removedAt,omitempty"`
}

type pageDTO struct {
	Pets       []petDTO `json:"pets"`
	NextCursor string   `json:"nextCursor"`
}

func encodePage(page domain.CatalogPage) ([]byte, error) {
	dto := pageDTO{NextCursor: page.NextCursor, Pets: make([]petDTO, len(page.Pets))}
	for i, pet := range page.Pets {
		dto.Pets[i] = toPetDTO(pet)
	}
	return json.Marshal(dto)
}

func decodePage(raw []byte) (domain.CatalogPage, error) {
	var dto pageDTO
	if err := json.Unmarshal(raw, &dto); err != nil {
		return domain.CatalogPage{}, err
	}
	pets := make([]domain.Pet, len(dto.Pets))
	for i, petDTO := range dto.Pets {
		pet, err := fromPetDTO(petDTO)
		if err != nil {
			return domain.CatalogPage{}, err
		}
		pets[i] = pet
	}
	return domain.CatalogPage{Pets: pets, NextCursor: dto.NextCursor}, nil
}

func toPetDTO(pet domain.Pet) petDTO {
	dto := petDTO{
		ID:               pet.ID.String(),
		StoreID:          pet.StoreID.String(),
		Name:             pet.Name,
		Species:          string(pet.Species),
		AgeYears:         pet.AgeYears,
		Description:      pet.Description,
		BreederName:      pet.BreederName,
		BreederEmail:     pet.BreederEmail,
		PictureObjectKey: pet.PictureObjectKey,
		Status:           string(pet.Status),
		CreatedAt:        pet.CreatedAt,
		SoldAt:           pet.SoldAt,
		RemovedAt:        pet.RemovedAt,
	}
	if pet.SoldByCustomerID != nil {
		customerID := pet.SoldByCustomerID.String()
		dto.SoldByCustomerID = &customerID
	}
	return dto
}

func fromPetDTO(dto petDTO) (domain.Pet, error) {
	id, err := uuid.Parse(dto.ID)
	if err != nil {
		return domain.Pet{}, fmt.Errorf("parse pet id: %w", err)
	}
	storeID, err := uuid.Parse(dto.StoreID)
	if err != nil {
		return domain.Pet{}, fmt.Errorf("parse store id: %w", err)
	}
	species, err := domain.ParseSpecies(dto.Species)
	if err != nil {
		return domain.Pet{}, err
	}
	status, err := domain.ParsePetStatus(dto.Status)
	if err != nil {
		return domain.Pet{}, err
	}
	var soldBy *uuid.UUID
	if dto.SoldByCustomerID != nil {
		customerID, err := uuid.Parse(*dto.SoldByCustomerID)
		if err != nil {
			return domain.Pet{}, fmt.Errorf("parse sold-by customer id: %w", err)
		}
		soldBy = &customerID
	}
	return domain.Pet{
		ID:               id,
		StoreID:          storeID,
		Name:             dto.Name,
		Species:          species,
		AgeYears:         dto.AgeYears,
		Description:      dto.Description,
		BreederName:      dto.BreederName,
		BreederEmail:     dto.BreederEmail,
		PictureObjectKey: dto.PictureObjectKey,
		Status:           status,
		CreatedAt:        dto.CreatedAt,
		SoldAt:           dto.SoldAt,
		SoldByCustomerID: soldBy,
		RemovedAt:        dto.RemovedAt,
	}, nil
}
