package graph

import (
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/google/uuid"
)

const soldPetsQuery = `query($from: Time!, $to: Time!) {
  soldPets(from: $from, to: $to, first: 10) { edges { node { id status } } }
}`

// The soldPets resolver path (Time-range parse, store scoping, sold-cursor
// mapping) is not reached by the app-layer test, so it is exercised here.
func TestE2E_SoldPets_ListsPurchasedWithinRange(t *testing.T) {
	requireInfra(t)
	storeID := seedStore(t)
	petID := createPetInStore(t, storeID)
	purchaseAs(t, client.New(handlerAs(ptr(customerIdentity(t)))), petID)

	merchant := client.New(handlerAs(ptr(merchantIdentity(storeID))))

	var out struct {
		SoldPets struct {
			Edges []struct {
				Node struct{ ID, Status string }
			}
		}
	}
	merchant.MustPost(soldPetsQuery, &out,
		client.Var("from", "2000-01-01T00:00:00Z"),
		client.Var("to", "2999-01-01T00:00:00Z"))

	if len(out.SoldPets.Edges) != 1 || out.SoldPets.Edges[0].Node.ID != petID {
		t.Fatalf("a wide range should return the sold pet, got %d edges", len(out.SoldPets.Edges))
	}
	if out.SoldPets.Edges[0].Node.Status != "SOLD" {
		t.Fatalf("status = %q, want SOLD", out.SoldPets.Edges[0].Node.Status)
	}

	var empty struct {
		SoldPets struct {
			Edges []struct {
				Node struct{ ID string }
			}
		}
	}
	merchant.MustPost(soldPetsQuery, &empty,
		client.Var("from", "2999-01-01T00:00:00Z"),
		client.Var("to", "3000-01-01T00:00:00Z"))
	if len(empty.SoldPets.Edges) != 0 {
		t.Fatalf("a future-only range must exclude the pet, got %d edges", len(empty.SoldPets.Edges))
	}
}

func TestE2E_SoldPets_StoreIsolation(t *testing.T) {
	requireInfra(t)
	storeA := seedStore(t)
	petID := createPetInStore(t, storeA)
	purchaseAs(t, client.New(handlerAs(ptr(customerIdentity(t)))), petID)

	otherMerchant := client.New(handlerAs(ptr(merchantIdentity(seedStore(t)))))
	var out struct {
		SoldPets struct {
			Edges []struct {
				Node struct{ ID string }
			}
		}
	}
	otherMerchant.MustPost(soldPetsQuery, &out,
		client.Var("from", "2000-01-01T00:00:00Z"),
		client.Var("to", "2999-01-01T00:00:00Z"))
	if len(out.SoldPets.Edges) != 0 {
		t.Fatalf("a merchant must not see another store's sold pets, got %d edges", len(out.SoldPets.Edges))
	}
}

// Every resolver that parses a client-supplied ID must reject a malformed one
// with VALIDATION before any service call.
func TestE2E_MalformedID_ReturnsValidation(t *testing.T) {
	requireInfra(t)
	storeID := seedStore(t)
	merchant := client.New(handlerAs(ptr(merchantIdentity(storeID))))
	customer := client.New(handlerAs(ptr(customerIdentity(t))))

	cases := []struct {
		name  string
		c     *client.Client
		query string
		opt   client.Option
	}{
		{"removePet", merchant, removePetMutation, client.Var("id", "not-a-uuid")},
		{"purchasePet", customer, purchasePetMutation, client.Var("petId", "not-a-uuid")},
		{"checkout", customer, checkoutMutation, client.Var("petIds", []string{"not-a-uuid"})},
		{"availablePets", customer, availablePetsQuery, client.Var("storeId", "not-a-uuid")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if code := errorCode(t, tc.c, tc.query, tc.opt); code != "VALIDATION" {
				t.Fatalf("%s with a malformed id: code = %q, want VALIDATION", tc.name, code)
			}
		})
	}
}

// A domain rule violation (here a negative age, which passes the GraphQL Int type
// but fails NewPet) must surface as a VALIDATION code, not INTERNAL.
func TestE2E_CreatePet_DomainValidation_ReturnsValidation(t *testing.T) {
	requireInfra(t)
	merchant := client.New(handlerAs(ptr(merchantIdentity(seedStore(t)))))

	input := createInput(t)
	input["ageYears"] = -1
	code := errorCode(t, merchant, createPetMutation, client.Var("input", input), client.WithFiles())
	if code != "VALIDATION" {
		t.Fatalf("createPet with a negative age: code = %q, want VALIDATION", code)
	}
}

// A nonexistent pet is NOT_FOUND, distinct from UNAVAILABLE (a real pet already sold).
func TestE2E_PurchasePet_Nonexistent_ReturnsNotFound(t *testing.T) {
	requireInfra(t)
	customer := client.New(handlerAs(ptr(customerIdentity(t))))
	if code := errorCode(t, customer, purchasePetMutation, client.Var("petId", uuid.New().String())); code != "NOT_FOUND" {
		t.Fatalf("purchasing a nonexistent pet: code = %q, want NOT_FOUND", code)
	}
}

// An empty cart succeeds with an empty result rather than erroring.
func TestE2E_Checkout_EmptyList_ReturnsEmpty(t *testing.T) {
	requireInfra(t)
	customer := client.New(handlerAs(ptr(customerIdentity(t))))

	var out struct {
		Checkout []struct{ ID string }
	}
	customer.MustPost(checkoutMutation, &out, client.Var("petIds", []string{}))
	if len(out.Checkout) != 0 {
		t.Fatalf("checkout with an empty cart should return no pets, got %d", len(out.Checkout))
	}
}
