package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"roboticCrewChallenge/internal/adapter/postgres/sqlcgen"
	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/platform/crypto"
	"roboticCrewChallenge/internal/platform/pagination"
)

type PetRepository struct {
	pool    *pgxpool.Pool
	queries *sqlcgen.Queries
	enc     *crypto.Encryptor
}

func NewPetRepository(pool *pgxpool.Pool, enc *crypto.Encryptor) *PetRepository {
	return &PetRepository{pool: pool, queries: sqlcgen.New(pool), enc: enc}
}

func (r *PetRepository) Create(ctx context.Context, pet domain.Pet) error {
	breederName, err := r.enc.Encrypt([]byte(pet.BreederName))
	if err != nil {
		return err
	}
	breederEmail, err := r.enc.Encrypt([]byte(pet.BreederEmail))
	if err != nil {
		return err
	}
	return r.queries.CreatePet(ctx, sqlcgen.CreatePetParams{
		ID:                    pet.ID,
		StoreID:               pet.StoreID,
		Name:                  pet.Name,
		Species:               string(pet.Species),
		AgeYears:              clampInt32(pet.AgeYears),
		Description:           pet.Description,
		BreederNameEncrypted:  breederName,
		BreederEmailEncrypted: breederEmail,
		PictureObjectKey:      pet.PictureObjectKey,
		Status:                string(pet.Status),
		CreatedAt:             pet.CreatedAt,
	})
}

func (r *PetRepository) GetByID(ctx context.Context, storeID, petID uuid.UUID) (domain.Pet, error) {
	row, err := r.queries.GetPetByID(ctx, sqlcgen.GetPetByIDParams{StoreID: storeID, ID: petID})
	if err != nil {
		return domain.Pet{}, mapNotFound(err, domain.ErrPetNotFound)
	}
	return r.toDomain(row)
}

func (r *PetRepository) Remove(ctx context.Context, storeID, petID uuid.UUID) (domain.Pet, error) {
	row, err := r.queries.RemovePet(ctx, sqlcgen.RemovePetParams{ID: petID, StoreID: storeID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Pet{}, r.removeFailure(ctx, storeID, petID)
		}
		return domain.Pet{}, err
	}
	return r.toDomain(row)
}

// removeFailure distinguishes "not in this store" from "no longer available".
func (r *PetRepository) removeFailure(ctx context.Context, storeID, petID uuid.UUID) error {
	if _, err := r.queries.GetPetByID(ctx, sqlcgen.GetPetByIDParams{StoreID: storeID, ID: petID}); err != nil {
		return mapNotFound(err, domain.ErrPetNotFound)
	}
	return domain.ErrPetNotRemovable
}

func (r *PetRepository) Purchase(ctx context.Context, customerID, petID uuid.UUID) (domain.Pet, error) {
	row, err := r.queries.PurchasePet(ctx, sqlcgen.PurchasePetParams{ID: petID, SoldByCustomerID: pgUUID(customerID)})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return r.purchaseFailure(ctx, customerID, petID)
		}
		return domain.Pet{}, err
	}
	return r.toDomain(row)
}

// purchaseFailure treats a pet already sold to the same customer as an idempotent success.
// The re-read runs outside the failed UPDATE's transaction, which is safe because SOLD is a
// terminal status: a sold pet never transitions again, so this verdict cannot go stale.
func (r *PetRepository) purchaseFailure(ctx context.Context, customerID, petID uuid.UUID) (domain.Pet, error) {
	row, err := r.queries.GetPetByIDUnscoped(ctx, petID)
	if err != nil {
		return domain.Pet{}, mapNotFound(err, domain.ErrPetNotFound)
	}
	pet, err := r.toDomain(row)
	if err != nil {
		return domain.Pet{}, err
	}
	if soldToSameCustomer(pet, customerID) {
		return pet, nil
	}
	return domain.Pet{}, domain.ErrPetUnavailable
}

func (r *PetRepository) Checkout(ctx context.Context, customerID uuid.UUID, petIDs []uuid.UUID) ([]domain.Pet, error) {
	petIDs = dedupeIDs(petIDs)

	var purchased []domain.Pet
	err := runInTx(ctx, r.pool, func(q *sqlcgen.Queries) error {
		locked, err := q.LockPetsByIDs(ctx, petIDs)
		if err != nil {
			return err
		}
		byID := make(map[uuid.UUID]domain.Pet, len(locked))
		for _, row := range locked {
			pet, err := r.toDomain(row)
			if err != nil {
				return err
			}
			byID[pet.ID] = pet
		}

		var unavailable []domain.UnavailablePet
		for _, id := range petIDs {
			pet, found := byID[id]
			switch {
			case !found:
				unavailable = append(unavailable, domain.UnavailablePet{ID: id})
			case pet.Status == domain.PetStatusAvailable, soldToSameCustomer(pet, customerID):
				// purchasable, or already owned by this customer
			default:
				unavailable = append(unavailable, domain.UnavailablePet{ID: pet.ID, Name: pet.Name})
			}
		}
		if len(unavailable) > 0 {
			return &domain.UnavailablePetsError{Pets: unavailable}
		}

		result := make([]domain.Pet, 0, len(petIDs))
		for _, id := range petIDs {
			if soldToSameCustomer(byID[id], customerID) {
				result = append(result, byID[id])
				continue
			}
			row, err := q.PurchasePet(ctx, sqlcgen.PurchasePetParams{ID: id, SoldByCustomerID: pgUUID(customerID)})
			if err != nil {
				return err
			}
			pet, err := r.toDomain(row)
			if err != nil {
				return err
			}
			result = append(result, pet)
		}
		purchased = result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return purchased, nil
}

func (r *PetRepository) ListAvailableByStore(ctx context.Context, storeID uuid.UUID, limit int, cursor string) ([]domain.Pet, string, error) {
	after, err := decodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	rows, err := r.queries.ListAvailableByStore(ctx, sqlcgen.ListAvailableByStoreParams{
		StoreID:        storeID,
		AfterCreatedAt: after.SortKey,
		AfterID:        after.ID,
		PageLimit:      clampInt32(limit),
	})
	if err != nil {
		return nil, "", err
	}
	pets, err := r.toDomainList(rows)
	if err != nil {
		return nil, "", err
	}
	return pets, nextCursor(pets, limit, func(p domain.Pet) pagination.Cursor {
		return pagination.Cursor{SortKey: p.CreatedAt, ID: p.ID}
	}), nil
}

func (r *PetRepository) ListSoldByStore(ctx context.Context, storeID uuid.UUID, from, to time.Time, limit int, cursor string) ([]domain.Pet, string, error) {
	after, err := decodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	rows, err := r.queries.ListSoldByStore(ctx, sqlcgen.ListSoldByStoreParams{
		StoreID:     storeID,
		SoldFrom:    pgTime(from),
		SoldTo:      pgTime(to),
		AfterSoldAt: pgTime(after.SortKey),
		AfterID:     after.ID,
		PageLimit:   clampInt32(limit),
	})
	if err != nil {
		return nil, "", err
	}
	pets, err := r.toDomainList(rows)
	if err != nil {
		return nil, "", err
	}
	return pets, nextCursor(pets, limit, func(p domain.Pet) pagination.Cursor {
		c := pagination.Cursor{ID: p.ID}
		if p.SoldAt != nil {
			c.SortKey = *p.SoldAt
		}
		return c
	}), nil
}

func (r *PetRepository) toDomainList(rows []sqlcgen.Pet) ([]domain.Pet, error) {
	pets := make([]domain.Pet, 0, len(rows))
	for _, row := range rows {
		pet, err := r.toDomain(row)
		if err != nil {
			return nil, err
		}
		pets = append(pets, pet)
	}
	return pets, nil
}

func (r *PetRepository) toDomain(row sqlcgen.Pet) (domain.Pet, error) {
	breederName, err := r.enc.Decrypt(row.BreederNameEncrypted)
	if err != nil {
		return domain.Pet{}, err
	}
	breederEmail, err := r.enc.Decrypt(row.BreederEmailEncrypted)
	if err != nil {
		return domain.Pet{}, err
	}
	species, err := domain.ParseSpecies(row.Species)
	if err != nil {
		return domain.Pet{}, err
	}
	status, err := domain.ParsePetStatus(row.Status)
	if err != nil {
		return domain.Pet{}, err
	}
	return domain.Pet{
		ID:               row.ID,
		StoreID:          row.StoreID,
		Name:             row.Name,
		Species:          species,
		AgeYears:         int(row.AgeYears),
		Description:      row.Description,
		BreederName:      string(breederName),
		BreederEmail:     string(breederEmail),
		PictureObjectKey: row.PictureObjectKey,
		Status:           status,
		CreatedAt:        row.CreatedAt,
		SoldAt:           timePtr(row.SoldAt),
		SoldByCustomerID: uuidPtr(row.SoldByCustomerID),
		RemovedAt:        timePtr(row.RemovedAt),
	}, nil
}

func soldToSameCustomer(pet domain.Pet, customerID uuid.UUID) bool {
	return pet.Status == domain.PetStatusSold &&
		pet.SoldByCustomerID != nil &&
		*pet.SoldByCustomerID == customerID
}

// dedupeIDs removes duplicate ids, preserving first-seen order, so a cart holding the
// same pet twice resolves to a single purchase rather than failing the whole checkout.
func dedupeIDs(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func decodeCursor(cursor string) (pagination.Cursor, error) {
	if cursor == "" {
		return pagination.Cursor{}, nil
	}
	return pagination.Decode(cursor)
}

func nextCursor(pets []domain.Pet, limit int, key func(domain.Pet) pagination.Cursor) string {
	if limit <= 0 || len(pets) < limit {
		return ""
	}
	return pagination.Encode(key(pets[len(pets)-1]))
}
