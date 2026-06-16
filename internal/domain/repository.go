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

// PictureStore persists pet pictures in object storage. Upload returns the
// generated object key the caller records on the pet; the picture bytes never
// touch the database. PresignedURL hands back a short-lived signed GET URL so
// clients fetch images directly from storage rather than through a resolver.
type PictureStore interface {
	Upload(ctx context.Context, body io.Reader, size int64, contentType string) (objectKey string, err error)
	PresignedURL(ctx context.Context, objectKey string) (string, error)
}
