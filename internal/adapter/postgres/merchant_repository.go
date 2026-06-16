package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"roboticCrewChallenge/internal/adapter/postgres/sqlcgen"
	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/platform/crypto"
)

type MerchantRepository struct {
	queries *sqlcgen.Queries
	enc     *crypto.Encryptor
	blind   *crypto.BlindIndex
}

func NewMerchantRepository(pool *pgxpool.Pool, enc *crypto.Encryptor, blind *crypto.BlindIndex) *MerchantRepository {
	return &MerchantRepository{queries: sqlcgen.New(pool), enc: enc, blind: blind}
}

func (r *MerchantRepository) Create(ctx context.Context, merchant domain.Merchant) error {
	emailEncrypted, err := r.enc.Encrypt([]byte(merchant.Email))
	if err != nil {
		return err
	}
	err = r.queries.CreateMerchant(ctx, sqlcgen.CreateMerchantParams{
		ID:             merchant.ID,
		EmailHash:      r.blind.Compute(merchant.Email),
		EmailEncrypted: emailEncrypted,
		PasswordHash:   merchant.PasswordHash,
		CreatedAt:      merchant.CreatedAt,
	})
	if isUniqueViolation(err) {
		return domain.ErrEmailInUse
	}
	return err
}

func (r *MerchantRepository) GetByEmail(ctx context.Context, email string) (domain.Merchant, error) {
	row, err := r.queries.GetMerchantByEmailHash(ctx, r.blind.Compute(email))
	if err != nil {
		return domain.Merchant{}, mapNotFound(err, domain.ErrMerchantNotFound)
	}
	return r.toDomain(row)
}

func (r *MerchantRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Merchant, error) {
	row, err := r.queries.GetMerchantByID(ctx, id)
	if err != nil {
		return domain.Merchant{}, mapNotFound(err, domain.ErrMerchantNotFound)
	}
	return r.toDomain(row)
}

func (r *MerchantRepository) toDomain(row sqlcgen.Merchant) (domain.Merchant, error) {
	email, err := r.enc.Decrypt(row.EmailEncrypted)
	if err != nil {
		return domain.Merchant{}, err
	}
	return domain.Merchant{
		ID:           row.ID,
		Email:        string(email),
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
	}, nil
}
