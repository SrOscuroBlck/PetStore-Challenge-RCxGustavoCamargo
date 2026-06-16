package domain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Msg)
}

var (
	ErrPetNotFound        = errors.New("pet not found")
	ErrPetUnavailable     = errors.New("pet is no longer available")
	ErrPetNotRemovable    = errors.New("pet cannot be removed")
	ErrStoreNotFound      = errors.New("store not found")
	ErrStoreAlreadyExists = errors.New("merchant already has a store")
	ErrMerchantNotFound   = errors.New("merchant not found")
	ErrCustomerNotFound   = errors.New("customer not found")
	ErrEmailInUse         = errors.New("email already in use")

	ErrUnsupportedPictureType = errors.New("unsupported picture content type")
	ErrPictureTooLarge        = errors.New("picture exceeds the maximum allowed size")
)

type UnavailablePet struct {
	ID   uuid.UUID
	Name string
}

type UnavailablePetsError struct {
	Pets []UnavailablePet
}

func (e *UnavailablePetsError) Error() string {
	names := make([]string, len(e.Pets))
	for i, pet := range e.Pets {
		if pet.Name != "" {
			names[i] = pet.Name
		} else {
			names[i] = pet.ID.String()
		}
	}
	return "these pets are no longer available: " + strings.Join(names, ", ")
}
