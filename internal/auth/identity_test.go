package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestIdentityContextRoundTrip(t *testing.T) {
	want := Identity{Subject: uuid.New(), Role: RoleMerchant, StoreID: uuid.New()}

	got, ok := IdentityFromContext(WithIdentity(context.Background(), want))

	if !ok {
		t.Fatal("expected an identity in context")
	}
	if got != want {
		t.Fatalf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestIdentityFromEmptyContext(t *testing.T) {
	if _, ok := IdentityFromContext(context.Background()); ok {
		t.Fatal("expected no identity in an empty context")
	}
}

func TestRoleValid(t *testing.T) {
	cases := map[Role]bool{
		RoleMerchant:  true,
		RoleCustomer:  true,
		Role("admin"): false,
		Role(""):      false,
	}
	for role, want := range cases {
		if got := role.Valid(); got != want {
			t.Errorf("Role(%q).Valid() = %v, want %v", role, got, want)
		}
	}
}
