package auth

import (
	"context"

	"github.com/google/uuid"
)

type Role string

const (
	RoleMerchant Role = "merchant"
	RoleCustomer Role = "customer"
)

func (r Role) Valid() bool {
	switch r {
	case RoleMerchant, RoleCustomer:
		return true
	default:
		return false
	}
}

// Identity is the authenticated principal for a request. StoreID is set only
// when Role is RoleMerchant and is the sole source of a merchant's store scope —
// it is never accepted from client input.
type Identity struct {
	Subject uuid.UUID
	Role    Role
	StoreID uuid.UUID
}

type contextKey struct{}

func WithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, contextKey{}, identity)
}

func IdentityFromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(contextKey{}).(Identity)
	return identity, ok
}
