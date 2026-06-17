package graph

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/google/uuid"
)

const (
	availablePetsQuery = `query($storeId: ID!) {
  availablePets(storeId: $storeId, first: 10) {
    edges { node { id status } cursor }
    pageInfo { hasNextPage endCursor }
  }
}`
	purchasePetMutation = `mutation($petId: ID!) { purchasePet(petId: $petId) { id status } }`
	checkoutMutation    = `mutation($petIds: [ID!]!) { checkout(petIds: $petIds) { id status } }`
)

func createPetInStore(t *testing.T, storeID uuid.UUID) string {
	t.Helper()
	merchant := client.New(handlerAs(ptr(merchantIdentity(storeID))))
	return createPet(t, merchant)
}

func purchaseAs(t *testing.T, c *client.Client, petID string) string {
	t.Helper()
	var out struct {
		PurchasePet struct{ ID, Status string }
	}
	c.MustPost(purchasePetMutation, &out, client.Var("petId", petID))
	return out.PurchasePet.Status
}

func errorMessage(t *testing.T, c *client.Client, query string, opts ...client.Option) string {
	t.Helper()
	resp, err := c.RawPost(query, opts...)
	if err != nil {
		t.Fatalf("raw post: %v", err)
	}
	var errs []struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(resp.Errors, &errs); err != nil {
		t.Fatalf("decode errors %q: %v", string(resp.Errors), err)
	}
	if len(errs) == 0 {
		t.Fatalf("expected a GraphQL error, got none")
	}
	return errs[0].Message
}

func availableIDs(t *testing.T, c *client.Client, storeID uuid.UUID) []string {
	t.Helper()
	var out struct {
		AvailablePets struct {
			Edges []struct {
				Node struct{ ID, Status string }
			}
		}
	}
	const query = `query($storeId: ID!) { availablePets(storeId: $storeId, first: 10) { edges { node { id status } } } }`
	c.MustPost(query, &out, client.Var("storeId", storeID.String()))
	ids := make([]string, 0, len(out.AvailablePets.Edges))
	for _, edge := range out.AvailablePets.Edges {
		if edge.Node.Status != "AVAILABLE" {
			t.Fatalf("availablePets returned a non-available pet: %s", edge.Node.Status)
		}
		ids = append(ids, edge.Node.ID)
	}
	return ids
}

func TestE2E_AvailablePets_ShowsOnlyAvailable(t *testing.T) {
	requireInfra(t)
	storeID := seedStore(t)
	petID := createPetInStore(t, storeID)
	customer := client.New(handlerAs(ptr(customerIdentity(t))))

	if ids := availableIDs(t, customer, storeID); len(ids) != 1 || ids[0] != petID {
		t.Fatalf("expected the created pet to be browsable, got %v", ids)
	}

	purchaseAs(t, client.New(handlerAs(ptr(customerIdentity(t)))), petID)

	if ids := availableIDs(t, customer, storeID); len(ids) != 0 {
		t.Fatalf("a purchased pet must no longer be browsable, got %v", ids)
	}
}

func TestE2E_PublicPet_HasNoBreederFields(t *testing.T) {
	requireInfra(t)
	storeID := seedStore(t)
	createPetInStore(t, storeID)
	customer := client.New(handlerAs(ptr(customerIdentity(t))))

	// breederName is not a field on PublicPet, so the query is rejected by schema
	// validation (HTTP 422) — the gqlgen client surfaces that as an error naming
	// the offending field.
	_, err := customer.RawPost(
		`query($storeId: ID!) { availablePets(storeId: $storeId, first: 1) { edges { node { breederName } } } }`,
		client.Var("storeId", storeID.String()),
	)
	if err == nil {
		t.Fatal("selecting breederName on PublicPet must be a schema validation error")
	}
	if !strings.Contains(err.Error(), "breederName") {
		t.Fatalf("validation error should name the offending field, got %v", err)
	}
}

func TestE2E_PurchasePet_SoldThenUnavailable(t *testing.T) {
	requireInfra(t)
	storeID := seedStore(t)
	petID := createPetInStore(t, storeID)

	if status := purchaseAs(t, client.New(handlerAs(ptr(customerIdentity(t)))), petID); status != "SOLD" {
		t.Fatalf("purchased pet status = %q, want SOLD", status)
	}

	other := client.New(handlerAs(ptr(customerIdentity(t))))
	if code := errorCode(t, other, purchasePetMutation, client.Var("petId", petID)); code != "UNAVAILABLE" {
		t.Fatalf("purchasing a sold pet: code = %q, want UNAVAILABLE", code)
	}
}

func TestE2E_Checkout_AtomicSuccess(t *testing.T) {
	requireInfra(t)
	storeID := seedStore(t)
	a := createPetInStore(t, storeID)
	b := createPetInStore(t, storeID)
	customer := client.New(handlerAs(ptr(customerIdentity(t))))

	var out struct {
		Checkout []struct{ ID, Status string }
	}
	customer.MustPost(checkoutMutation, &out, client.Var("petIds", []string{a, b}))
	if len(out.Checkout) != 2 {
		t.Fatalf("expected 2 purchased pets, got %d", len(out.Checkout))
	}
	for _, pet := range out.Checkout {
		if pet.Status != "SOLD" {
			t.Fatalf("checked-out pet status = %q, want SOLD", pet.Status)
		}
	}
}

func TestE2E_Checkout_UnavailableNamesPet(t *testing.T) {
	requireInfra(t)
	storeID := seedStore(t)
	available := createPetInStore(t, storeID)
	taken := createPetInStore(t, storeID)
	purchaseAs(t, client.New(handlerAs(ptr(customerIdentity(t)))), taken)

	customer := client.New(handlerAs(ptr(customerIdentity(t))))
	petIDs := []string{available, taken}
	if msg := errorMessage(t, customer, checkoutMutation, client.Var("petIds", petIDs)); !strings.Contains(msg, "Pluto") {
		t.Fatalf("checkout error must name the unavailable pet, got %q", msg)
	}
	if code := errorCode(t, customer, checkoutMutation, client.Var("petIds", petIDs)); code != "UNAVAILABLE" {
		t.Fatalf("checkout-with-unavailable code = %q, want UNAVAILABLE", code)
	}
}

func TestE2E_RoleSeparation_MerchantCannotUseCustomerOps(t *testing.T) {
	requireInfra(t)
	storeID := seedStore(t)
	petID := createPetInStore(t, storeID)
	merchant := client.New(handlerAs(ptr(merchantIdentity(storeID))))

	if code := errorCode(t, merchant, purchasePetMutation, client.Var("petId", petID)); code != "FORBIDDEN" {
		t.Fatalf("merchant calling purchasePet: code = %q, want FORBIDDEN", code)
	}
	if code := errorCode(t, merchant, availablePetsQuery, client.Var("storeId", storeID.String())); code != "FORBIDDEN" {
		t.Fatalf("merchant calling availablePets: code = %q, want FORBIDDEN", code)
	}
}
