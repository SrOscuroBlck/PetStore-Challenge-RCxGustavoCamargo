package rediscache

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/platform/crypto"
)

type Cache struct {
	client *redis.Client
	enc    *crypto.Encryptor
	logger *slog.Logger
	ttl    time.Duration
}

func New(client *redis.Client, enc *crypto.Encryptor, logger *slog.Logger, ttl time.Duration) *Cache {
	return &Cache{client: client, enc: enc, logger: logger, ttl: ttl}
}

var _ domain.CatalogCache = (*Cache)(nil)

func (c *Cache) GetAvailable(ctx context.Context, storeID uuid.UUID, species *domain.Species, limit int, cursor string) (domain.CatalogPage, bool) {
	generation, ok := c.generation(ctx, storeID)
	if !ok {
		return domain.CatalogPage{}, false
	}
	raw, err := c.client.Get(ctx, pageKey(storeID, generation, species, limit, cursor)).Bytes()
	if errors.Is(err, redis.Nil) {
		return domain.CatalogPage{}, false
	}
	if err != nil {
		c.logger.WarnContext(ctx, "catalog cache read failed", "store_id", storeID, "error", err)
		return domain.CatalogPage{}, false
	}
	plaintext, err := c.enc.Decrypt(raw)
	if err != nil {
		c.logger.WarnContext(ctx, "catalog cache decrypt failed", "store_id", storeID, "error", err)
		return domain.CatalogPage{}, false
	}
	page, err := decodePage(plaintext)
	if err != nil {
		c.logger.WarnContext(ctx, "catalog cache decode failed", "store_id", storeID, "error", err)
		return domain.CatalogPage{}, false
	}
	return page, true
}

func (c *Cache) SetAvailable(ctx context.Context, storeID uuid.UUID, species *domain.Species, limit int, cursor string, page domain.CatalogPage) {
	generation, ok := c.generation(ctx, storeID)
	if !ok {
		return
	}
	plaintext, err := encodePage(page)
	if err != nil {
		c.logger.ErrorContext(ctx, "catalog cache encode failed", "store_id", storeID, "error", err)
		return
	}
	ciphertext, err := c.enc.Encrypt(plaintext)
	if err != nil {
		c.logger.ErrorContext(ctx, "catalog cache encrypt failed", "store_id", storeID, "error", err)
		return
	}
	if err := c.client.Set(ctx, pageKey(storeID, generation, species, limit, cursor), ciphertext, c.ttl).Err(); err != nil {
		c.logger.WarnContext(ctx, "catalog cache write failed", "store_id", storeID, "error", err)
	}
}

func (c *Cache) InvalidateStore(ctx context.Context, storeID uuid.UUID) {
	if err := c.client.Incr(ctx, genKey(storeID)).Err(); err != nil {
		c.logger.WarnContext(ctx, "catalog cache invalidate failed", "store_id", storeID, "error", err)
	}
}

// generation returns ok=false on a Redis error so the caller degrades to a miss
// or no-op rather than failing the request; a never-set counter is generation zero.
func (c *Cache) generation(ctx context.Context, storeID uuid.UUID) (int64, bool) {
	generation, err := c.client.Get(ctx, genKey(storeID)).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, true
	}
	if err != nil {
		c.logger.WarnContext(ctx, "catalog cache generation read failed", "store_id", storeID, "error", err)
		return 0, false
	}
	return generation, true
}
