package domain

import "fmt"

type PetStatus string

const (
	PetStatusAvailable PetStatus = "AVAILABLE"
	PetStatusSold      PetStatus = "SOLD"
	PetStatusRemoved   PetStatus = "REMOVED"
)

func ParsePetStatus(value string) (PetStatus, error) {
	status := PetStatus(value)
	if !status.Valid() {
		return "", &ValidationError{Field: "status", Msg: fmt.Sprintf("unknown status %q", value)}
	}
	return status, nil
}

func (s PetStatus) Valid() bool {
	switch s {
	case PetStatusAvailable, PetStatusSold, PetStatusRemoved:
		return true
	default:
		return false
	}
}
