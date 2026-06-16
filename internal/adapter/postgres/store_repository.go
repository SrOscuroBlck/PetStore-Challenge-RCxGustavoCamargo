package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"roboticCrewChallenge/internal/adapter/postgres/sqlcgen"
	"roboticCrewChallenge/internal/domain"
)

type StoreRepository struct {
	queries *sqlcgen.Queries
}

func NewStoreRepository(pool *pgxpool.Pool) *StoreRepository {
	return &StoreRepository{queries: sqlcgen.New(pool)}
}

func (r *StoreRepository) Create(ctx context.Context, store domain.Store) error {
	err := r.queries.CreateStore(ctx, sqlcgen.CreateStoreParams{
		ID:         store.ID,
		MerchantID: store.MerchantID,
		Name:       store.Name,
		CreatedAt:  store.CreatedAt,
	})
	if isUniqueViolation(err) {
		return domain.ErrStoreAlreadyExists
	}
	return err
}

func (r *StoreRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) (domain.Store, error) {
	row, err := r.queries.GetStoreByMerchantID(ctx, merchantID)
	if err != nil {
		return domain.Store{}, mapNotFound(err, domain.ErrStoreNotFound)
	}
	return storeToDomain(row), nil
}

func (r *StoreRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Store, error) {
	row, err := r.queries.GetStoreByID(ctx, id)
	if err != nil {
		return domain.Store{}, mapNotFound(err, domain.ErrStoreNotFound)
	}
	return storeToDomain(row), nil
}

func storeToDomain(row sqlcgen.Store) domain.Store {
	return domain.Store{
		ID:         row.ID,
		MerchantID: row.MerchantID,
		Name:       row.Name,
		CreatedAt:  row.CreatedAt,
	}
}
