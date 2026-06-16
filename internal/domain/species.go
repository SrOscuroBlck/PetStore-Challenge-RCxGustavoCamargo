package domain

import "fmt"

type Species string

const (
	SpeciesCat  Species = "CAT"
	SpeciesDog  Species = "DOG"
	SpeciesFrog Species = "FROG"
)

func ParseSpecies(value string) (Species, error) {
	species := Species(value)
	if !species.Valid() {
		return "", &ValidationError{Field: "species", Msg: fmt.Sprintf("unknown species %q", value)}
	}
	return species, nil
}

func (s Species) Valid() bool {
	switch s {
	case SpeciesCat, SpeciesDog, SpeciesFrog:
		return true
	default:
		return false
	}
}
