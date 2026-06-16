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
	ErrPetNotFound      = errors.New("pet not found")
	ErrPetUnavailable   = errors.New("pet is no longer available")
	ErrPetNotRemovable  = errors.New("pet cannot be removed")
	ErrStoreNotFound    = errors.New("store not found")
	ErrMerchantNotFound = errors.New("merchant not found")
	ErrCustomerNotFound = errors.New("customer not found")
	ErrEmailInUse       = errors.New("email already in use")
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
		names[i] = pet.Name
	}
	return "these pets are no longer available: " + strings.Join(names, ", ")
}
