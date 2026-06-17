package graph

import (
	"roboticCrewChallenge/internal/app/listing"
	"roboticCrewChallenge/internal/app/purchase"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// Resolver is the dependency root injected into every resolver.
type Resolver struct {
	Listing  *listing.Service
	Purchase *purchase.Service
}

func clampFirst(first *int) int {
	if first == nil || *first <= 0 {
		return defaultPageSize
	}
	if *first > maxPageSize {
		return maxPageSize
	}
	return *first
}

func cursorOrEmpty(after *string) string {
	if after == nil {
		return ""
	}
	return *after
}
