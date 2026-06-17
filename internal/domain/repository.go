package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

type PetRepository interface {
	Create(ctx context.Context, pet Pet) error
	GetByID(ctx context.Context, storeID, petID uuid.UUID) (Pet, error)
	Remove(ctx context.Context, storeID, petID uuid.UUID) (Pet, error)
	ListAvailableByStore(ctx context.Context, storeID uuid.UUID, limit int, cursor string) ([]Pet, string, error)
	ListSoldByStore(ctx context.Context, storeID uuid.UUID, from, to time.Time, limit int, cursor string) ([]Pet, string, error)
	Purchase(ctx context.Context, customerID, petID uuid.UUID) (Pet, error)
	Checkout(ctx context.Context, customerID uuid.UUID, petIDs []uuid.UUID) ([]Pet, error)
}

type MerchantRepository interface {
	Create(ctx context.Context, merchant Merchant) error
	GetByEmail(ctx context.Context, email string) (Merchant, error)
	GetByID(ctx context.Context, id uuid.UUID) (Merchant, error)
}

type StoreRepository interface {
	Create(ctx context.Context, store Store) error
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) (Store, error)
	GetByID(ctx context.Context, id uuid.UUID) (Store, error)
}

type CustomerRepository interface {
	Create(ctx context.Context, customer Customer) error
	GetByEmail(ctx context.Context, email string) (Customer, error)
	GetByID(ctx context.Context, id uuid.UUID) (Customer, error)
}

// PictureContent is a pet picture read back from object storage: a streamable
// body the caller must close, plus the metadata needed to serve it. Size is -1
// when the store cannot report it up front.
type PictureContent struct {
	Body        io.ReadCloser
	ContentType string
	Size        int64
}

// PictureStore persists pet pictures in object storage. Upload returns the
// generated object key the caller records on the pet; the picture bytes never
// touch the database. Get streams a stored picture back so it can be served
// through the API's picture path rather than loaded through a resolver.
type PictureStore interface {
	Upload(ctx context.Context, body io.Reader, size int64, contentType string) (objectKey string, err error)
	Get(ctx context.Context, objectKey string) (PictureContent, error)
}

// CatalogPage is one page of a store's available-pets listing together with the
// cursor to the next page — the unit a CatalogCache stores and returns.
type CatalogPage struct {
	Pets       []Pet
	NextCursor string
}

// CatalogCache caches a store's available-pets listing per page (cache-aside).
// It is an accelerator only: every method degrades to a miss or no-op when the
// backing store is unavailable, so callers always fall back to the source of
// truth. The implementation owns the key schema and invalidation mechanism;
// callers pass the logical coordinates (store, page) and never raw keys. Any
// operation that changes a store's available pets — create, remove, and a sale
// (the purchase use case in a later issue) — must call InvalidateStore.
type CatalogCache interface {
	GetAvailable(ctx context.Context, storeID uuid.UUID, limit int, cursor string) (CatalogPage, bool)
	SetAvailable(ctx context.Context, storeID uuid.UUID, limit int, cursor string, page CatalogPage)
	InvalidateStore(ctx context.Context, storeID uuid.UUID)
}
