package rediscache

import (
	"context"

	"github.com/google/uuid"

	"roboticCrewChallenge/internal/domain"
)

// NoOp is a CatalogCache that never caches: reads always miss and writes are
// dropped. It lets callers that do not exercise caching construct the listing
// service without a Redis connection.
type NoOp struct{}

var _ domain.CatalogCache = NoOp{}

func (NoOp) GetAvailable(context.Context, uuid.UUID, *domain.Species, int, string) (domain.CatalogPage, bool) {
	return domain.CatalogPage{}, false
}

func (NoOp) SetAvailable(context.Context, uuid.UUID, *domain.Species, int, string, domain.CatalogPage) {
}

func (NoOp) InvalidateStore(context.Context, uuid.UUID) {}
