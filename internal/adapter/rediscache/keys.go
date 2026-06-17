package rediscache

import (
	"fmt"

	"github.com/google/uuid"

	"roboticCrewChallenge/internal/domain"
)

const (
	// emptyCursorToken stands in for the first-page cursor so a key never ends bare.
	emptyCursorToken = "_"
	// allSpeciesToken keys the unfiltered listing so it never collides with a
	// species-filtered page.
	allSpeciesToken = "all"
)

func genKey(storeID uuid.UUID) string {
	return fmt.Sprintf("catalog:%s:gen", storeID)
}

func pageKey(storeID uuid.UUID, generation int64, species *domain.Species, limit int, cursor string) string {
	if cursor == "" {
		cursor = emptyCursorToken
	}
	speciesToken := allSpeciesToken
	if species != nil {
		speciesToken = string(*species)
	}
	return fmt.Sprintf("catalog:%s:g%d:sp%s:l%d:c%s", storeID, generation, speciesToken, limit, cursor)
}
