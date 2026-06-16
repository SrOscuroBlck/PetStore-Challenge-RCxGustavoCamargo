package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"roboticCrewChallenge/internal/adapter/postgres/sqlcgen"
	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/platform/crypto"
)

type CustomerRepository struct {
	queries *sqlcgen.Queries
	enc     *crypto.Encryptor
	blind   *crypto.BlindIndex
}

func NewCustomerRepository(pool *pgxpool.Pool, enc *crypto.Encryptor, blind *crypto.BlindIndex) *CustomerRepository {
	return &CustomerRepository{queries: sqlcgen.New(pool), enc: enc, blind: blind}
}

func (r *CustomerRepository) Create(ctx context.Context, customer domain.Customer) error {
	emailEncrypted, err := r.enc.Encrypt([]byte(customer.Email))
	if err != nil {
		return err
	}
	err = r.queries.CreateCustomer(ctx, sqlcgen.CreateCustomerParams{
		ID:             customer.ID,
		EmailHash:      r.blind.Compute(customer.Email),
		EmailEncrypted: emailEncrypted,
		PasswordHash:   customer.PasswordHash,
		CreatedAt:      customer.CreatedAt,
	})
	if isUniqueViolation(err) {
		return domain.ErrEmailInUse
	}
	return err
}

func (r *CustomerRepository) GetByEmail(ctx context.Context, email string) (domain.Customer, error) {
	row, err := r.queries.GetCustomerByEmailHash(ctx, r.blind.Compute(email))
	if err != nil {
		return domain.Customer{}, mapNotFound(err, domain.ErrCustomerNotFound)
	}
	return r.toDomain(row)
}

func (r *CustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Customer, error) {
	row, err := r.queries.GetCustomerByID(ctx, id)
	if err != nil {
		return domain.Customer{}, mapNotFound(err, domain.ErrCustomerNotFound)
	}
	return r.toDomain(row)
}

func (r *CustomerRepository) toDomain(row sqlcgen.Customer) (domain.Customer, error) {
	email, err := r.enc.Decrypt(row.EmailEncrypted)
	if err != nil {
		return domain.Customer{}, err
	}
	return domain.Customer{
		ID:           row.ID,
		Email:        string(email),
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
	}, nil
}
