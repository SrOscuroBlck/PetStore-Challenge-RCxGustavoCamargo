package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/platform/crypto"
)

// credentialCacheTTL bounds how long a verified credential is trusted without
// re-running bcrypt. Short enough that a revoked credential stops working
// promptly, long enough that a burst of requests from one principal pays the
// bcrypt cost once rather than per request.
const credentialCacheTTL = 60 * time.Second

type merchantLookup interface {
	GetByEmail(ctx context.Context, email string) (domain.Merchant, error)
}

type customerLookup interface {
	GetByEmail(ctx context.Context, email string) (domain.Customer, error)
}

type storeLookup interface {
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) (domain.Store, error)
}

type Authenticator struct {
	merchants merchantLookup
	customers customerLookup
	stores    storeLookup
	cache     *credentialCache
}

func NewAuthenticator(merchants merchantLookup, customers customerLookup, stores storeLookup) *Authenticator {
	return &Authenticator{
		merchants: merchants,
		customers: customers,
		stores:    stores,
		cache:     newCredentialCache(credentialCacheTTL),
	}
}

// Authenticate resolves email+password to an Identity, serving a recent success
// from the credential cache to avoid re-running bcrypt on every request. Only
// successful resolutions are cached; failures always pay the full cost.
func (a *Authenticator) Authenticate(ctx context.Context, email, password string) (Identity, error) {
	key := credentialKey(email, password)
	if identity, ok := a.cache.get(key); ok {
		return identity, nil
	}

	identity, err := a.resolve(ctx, email, password)
	if err != nil {
		return Identity{}, err
	}

	a.cache.put(key, identity)
	return identity, nil
}

// resolve performs the authoritative credential check. Merchants take precedence
// over customers when an email exists in both tables, which are assumed disjoint.
// Every credential failure collapses to ErrInvalidCredentials so callers cannot
// tell an unknown email from a wrong password; an unknown email still pays a
// bcrypt comparison (see equalizeTiming) so it is not distinguishable by the
// dominant timing cost. Infrastructure failures bubble up distinct from
// ErrInvalidCredentials so they are not mistaken for a 401.
func (a *Authenticator) resolve(ctx context.Context, email, password string) (Identity, error) {
	merchant, err := a.merchants.GetByEmail(ctx, email)
	switch {
	case err == nil:
		return a.authenticateMerchant(ctx, merchant, password)
	case !errors.Is(err, domain.ErrMerchantNotFound):
		return Identity{}, fmt.Errorf("lookup merchant: %w", err)
	}

	customer, err := a.customers.GetByEmail(ctx, email)
	switch {
	case err == nil:
		return a.authenticateCustomer(customer, password)
	case !errors.Is(err, domain.ErrCustomerNotFound):
		return Identity{}, fmt.Errorf("lookup customer: %w", err)
	}

	equalizeTiming(password)
	return Identity{}, ErrInvalidCredentials
}

func (a *Authenticator) authenticateMerchant(ctx context.Context, merchant domain.Merchant, password string) (Identity, error) {
	if err := crypto.VerifyPassword(merchant.PasswordHash, password); err != nil {
		return Identity{}, ErrInvalidCredentials
	}
	store, err := a.stores.GetByMerchantID(ctx, merchant.ID)
	switch {
	case errors.Is(err, domain.ErrStoreNotFound):
		return Identity{}, ErrInvalidCredentials
	case err != nil:
		return Identity{}, fmt.Errorf("load merchant store: %w", err)
	}
	return Identity{Subject: merchant.ID, Role: RoleMerchant, StoreID: store.ID}, nil
}

func (a *Authenticator) authenticateCustomer(customer domain.Customer, password string) (Identity, error) {
	if err := crypto.VerifyPassword(customer.PasswordHash, password); err != nil {
		return Identity{}, ErrInvalidCredentials
	}
	return Identity{Subject: customer.ID, Role: RoleCustomer}, nil
}

// equalizeTiming runs a bcrypt comparison against a throwaway hash so the
// unknown-email path costs roughly the same as a wrong-password path, closing a
// timing oracle that would otherwise reveal whether an account exists.
func equalizeTiming(password string) {
	_ = crypto.VerifyPassword(dummyPasswordHash(), password)
}

var dummyPasswordHash = sync.OnceValue(func() string {
	hash, err := crypto.HashPassword("timing-equalization placeholder")
	if err != nil {
		panic(fmt.Sprintf("auth: precompute dummy password hash: %v", err))
	}
	return hash
})
