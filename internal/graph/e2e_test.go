package graph

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/google/uuid"

	"roboticCrewChallenge/internal/adapter/postgres"
	"roboticCrewChallenge/internal/auth"
)

const createPetMutation = `mutation($input: CreatePetInput!) {
  createPet(input: $input) { id status createdAt pictureUrl }
}`

const removePetMutation = `mutation($id: ID!) { removePet(id: $id) { id status } }`

func createInput(t *testing.T) map[string]any {
	return map[string]any{
		"name":         "Pluto",
		"species":      "DOG",
		"ageYears":     3,
		"description":  "Friendly",
		"breederName":  "Jane Doe",
		"breederEmail": "jane@example.com",
		"picture":      writePNG(t),
	}
}

func createPet(t *testing.T, c *client.Client) string {
	t.Helper()
	var out struct {
		CreatePet struct {
			ID, Status, CreatedAt, PictureURL string
		}
	}
	c.MustPost(createPetMutation, &out, client.Var("input", createInput(t)), client.WithFiles())
	if out.CreatePet.ID == "" {
		t.Fatal("createPet returned no id")
	}
	return out.CreatePet.ID
}

func errorCode(t *testing.T, c *client.Client, query string, opts ...client.Option) string {
	t.Helper()
	resp, err := c.RawPost(query, opts...)
	if err != nil {
		t.Fatalf("raw post: %v", err)
	}
	var errs []struct {
		Message    string         `json:"message"`
		Extensions map[string]any `json:"extensions"`
	}
	if err := json.Unmarshal(resp.Errors, &errs); err != nil {
		t.Fatalf("decode errors %q: %v", string(resp.Errors), err)
	}
	if len(errs) == 0 {
		t.Fatalf("expected a GraphQL error, got none")
	}
	code, _ := errs[0].Extensions["code"].(string)
	return code
}

func TestE2E_CreatePet_ReturnsPetWithPictureURL(t *testing.T) {
	requireInfra(t)
	c := client.New(handlerAs(ptr(merchantIdentity(seedStore(t)))))

	var out struct {
		CreatePet struct {
			ID, Status, CreatedAt, PictureURL string
		}
	}
	c.MustPost(createPetMutation, &out, client.Var("input", createInput(t)), client.WithFiles())

	if out.CreatePet.Status != "AVAILABLE" {
		t.Fatalf("status = %q, want AVAILABLE", out.CreatePet.Status)
	}
	if out.CreatePet.CreatedAt == "" {
		t.Fatal("createdAt must be returned")
	}
	if out.CreatePet.PictureURL == "" {
		t.Fatal("pictureUrl must resolve to a presigned URL")
	}
}

func TestE2E_UnsoldPets_ListsCreatedPet(t *testing.T) {
	requireInfra(t)
	c := client.New(handlerAs(ptr(merchantIdentity(seedStore(t)))))
	id := createPet(t, c)

	var out struct {
		UnsoldPets struct {
			Edges []struct {
				Node   struct{ ID, Status string }
				Cursor string
			}
			PageInfo struct {
				HasNextPage bool
				EndCursor   *string
			}
		}
	}
	c.MustPost(`query { unsoldPets(first: 10) { edges { node { id status } cursor } pageInfo { hasNextPage endCursor } } }`, &out)

	if len(out.UnsoldPets.Edges) != 1 || out.UnsoldPets.Edges[0].Node.ID != id {
		t.Fatalf("expected the created pet in the connection, got %d edges", len(out.UnsoldPets.Edges))
	}
	if out.UnsoldPets.Edges[0].Cursor == "" {
		t.Fatal("edge cursor must be set")
	}
}

func TestE2E_RemoveAlreadySold_ReturnsConflict(t *testing.T) {
	requireInfra(t)
	storeID := seedStore(t)
	c := client.New(handlerAs(ptr(merchantIdentity(storeID))))
	id := createPet(t, c)

	customerID := seedCustomer(t)
	petID := uuid.MustParse(id)
	if _, err := postgres.NewPetRepository(testPool, testEnc).Purchase(context.Background(), customerID, petID); err != nil {
		t.Fatalf("purchase pet: %v", err)
	}

	if code := errorCode(t, c, removePetMutation, client.Var("id", id)); code != "CONFLICT" {
		t.Fatalf("error code = %q, want CONFLICT", code)
	}
}

func TestE2E_StoreIsolation_RemoveOtherStorePet_NotFound(t *testing.T) {
	requireInfra(t)
	merchantA := client.New(handlerAs(ptr(merchantIdentity(seedStore(t)))))
	id := createPet(t, merchantA)

	merchantB := client.New(handlerAs(ptr(merchantIdentity(seedStore(t)))))
	if code := errorCode(t, merchantB, removePetMutation, client.Var("id", id)); code != "NOT_FOUND" {
		t.Fatalf("error code = %q, want NOT_FOUND", code)
	}
}

func TestE2E_Unauthenticated_ReturnsUnauthenticated(t *testing.T) {
	requireInfra(t)
	c := client.New(handlerAs(nil))
	if code := errorCode(t, c, removePetMutation, client.Var("id", uuid.New().String())); code != "UNAUTHENTICATED" {
		t.Fatalf("error code = %q, want UNAUTHENTICATED", code)
	}
}

func TestE2E_CustomerRole_ReturnsForbidden(t *testing.T) {
	requireInfra(t)
	customer := auth.Identity{Subject: uuid.New(), Role: auth.RoleCustomer}
	c := client.New(handlerAs(&customer))
	if code := errorCode(t, c, removePetMutation, client.Var("id", uuid.New().String())); code != "FORBIDDEN" {
		t.Fatalf("error code = %q, want FORBIDDEN", code)
	}
}

func ptr(id auth.Identity) *auth.Identity { return &id }
