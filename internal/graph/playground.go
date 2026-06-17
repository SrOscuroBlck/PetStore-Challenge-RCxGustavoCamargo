package graph

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/playground"
)

// The playground is dev-only (mounted only when introspection is enabled). It is
// pre-seeded with one ready tab per operation, each carrying the matching demo
// credential, so a reviewer can run everything without typing headers or queries.
// These are the published demo credentials (see the README), not real secrets.
const (
	playgroundMerchantAuth  = "Basic bWVyY2hhbnRAcGV0c3RvcmUubG9jYWw6ZGVtby1wYXNzd29yZA=="
	playgroundMerchant2Auth = "Basic bWVyY2hhbnQyQHBldHN0b3JlLmxvY2FsOmRlbW8tcGFzc3dvcmQ="
	playgroundCustomerAuth  = "Basic Y3VzdG9tZXJAcGV0c3RvcmUubG9jYWw6ZGVtby1wYXNzd29yZA=="
	playgroundCustomer2Auth = "Basic Y3VzdG9tZXIyQHBldHN0b3JlLmxvY2FsOmRlbW8tcGFzc3dvcmQ="
	playgroundStoreID       = "11111111-1111-1111-1111-111111111111"
)

// NewPlaygroundHandler serves an Altair playground pre-loaded with the customer
// and merchant operations. Altair runs in the browser same-origin with the API,
// so it reuses the TLS cert the browser already trusts (no separate SSL setup).
func NewPlaygroundHandler(endpoint string) http.Handler {
	return playground.AltairHandler("Pet Store API", endpoint, map[string]any{
		"initialWindows": playgroundWindows(),
	})
}

func playgroundWindow(name, query, auth, variables string) map[string]any {
	return map[string]any{
		"initialName":      name,
		"initialQuery":     query,
		"initialHeaders":   map[string]string{"Authorization": auth},
		"initialVariables": variables,
	}
}

func playgroundWindows() []map[string]any {
	return []map[string]any{
		playgroundWindow("M1 · createPet (add a file variable)",
			"# In the Variables panel, click the file icon, add a variable named\n"+
				"# \"picture\", and pick a JPEG/PNG — then Send.\n"+
				"mutation CreatePet($picture: Upload!) {\n"+
				"  createPet(input: {\n"+
				"    name: \"Biscuit\", species: DOG, ageYears: 2,\n"+
				"    description: \"A cheerful rescue pup.\",\n"+
				"    breederName: \"Jane Doe\", breederEmail: \"jane@example.com\",\n"+
				"    picture: $picture\n"+
				"  }) { id name species ageYears description breederName breederEmail pictureUrl status createdAt }\n"+
				"}\n",
			playgroundMerchantAuth, "{}"),

		playgroundWindow("M2 · unsoldPets (full details)",
			"query UnsoldPets($first: Int, $after: String) {\n"+
				"  unsoldPets(first: $first, after: $after) {\n"+
				"    edges { node { id name species ageYears description breederName breederEmail pictureUrl status createdAt } cursor }\n"+
				"    pageInfo { hasNextPage endCursor }\n"+
				"  }\n}\n",
			playgroundMerchantAuth, "{\n  \"first\": 12,\n  \"after\": null\n}"),

		playgroundWindow("M3 · soldPets (inclusive range, full details)",
			"query SoldPets($from: Time!, $to: Time!, $first: Int) {\n"+
				"  soldPets(from: $from, to: $to, first: $first) {\n"+
				"    edges { node { id name species ageYears description breederName breederEmail pictureUrl status createdAt soldAt } cursor }\n"+
				"    pageInfo { hasNextPage endCursor }\n"+
				"  }\n}\n",
			playgroundMerchantAuth, "{\n  \"from\": \"2000-01-01T00:00:00Z\",\n  \"to\": \"2999-01-01T00:00:00Z\",\n  \"first\": 12\n}"),

		playgroundWindow("M4 · removePet (paste an AVAILABLE id)",
			"mutation RemovePet($id: ID!) {\n  removePet(id: $id) { id name status }\n}\n",
			playgroundMerchantAuth, "{\n  \"id\": \"PASTE-AVAILABLE-PET-ID\"\n}"),

		playgroundWindow("M5 · role separation → FORBIDDEN",
			"# Auth here is the MERCHANT; a customer op must be rejected.\n"+
				"query { availablePets(storeId: \""+playgroundStoreID+"\", first: 1) { edges { node { id } } } }\n",
			playgroundMerchantAuth, "{}"),

		playgroundWindow("M6 · store isolation (merchant 2 sees only its store)",
			"# Auth here is MERCHANT 2 — returns only the Second Store's pets.\n"+
				"query { unsoldPets(first: 50) { edges { node { id name species } } } }\n",
			playgroundMerchant2Auth, "{}"),

		playgroundWindow("M7 · store isolation (merchant 2 → store-1 id = NOT_FOUND)",
			"# Auth is MERCHANT 2; paste a STORE-1 pet id → must be NOT_FOUND.\n"+
				"mutation RemovePet($id: ID!) {\n  removePet(id: $id) { id status }\n}\n",
			playgroundMerchant2Auth, "{\n  \"id\": \"PASTE-A-STORE-1-PET-ID\"\n}"),

		playgroundWindow("C1 · availablePets (optional species filter)",
			"query AvailablePets($storeId: ID!, $species: Species, $first: Int, $after: String) {\n"+
				"  availablePets(storeId: $storeId, species: $species, first: $first, after: $after) {\n"+
				"    edges { node { id name species ageYears description pictureUrl status createdAt } cursor }\n"+
				"    pageInfo { hasNextPage endCursor }\n"+
				"  }\n}\n",
			playgroundCustomerAuth, "{\n  \"storeId\": \""+playgroundStoreID+"\",\n  \"species\": null,\n  \"first\": 12,\n  \"after\": null\n}"),

		playgroundWindow("C2 · purchasePet (paste an AVAILABLE id)",
			"mutation PurchasePet($petId: ID!) {\n  purchasePet(petId: $petId) { id name status }\n}\n",
			playgroundCustomerAuth, "{\n  \"petId\": \"PASTE-AVAILABLE-PET-ID\"\n}"),

		playgroundWindow("C3 · checkout (paste 2–3 AVAILABLE ids)",
			"mutation Checkout($petIds: [ID!]!) {\n  checkout(petIds: $petIds) { id name status }\n}\n",
			playgroundCustomerAuth, "{\n  \"petIds\": [\"PASTE-ID-1\", \"PASTE-ID-2\"]\n}"),

		playgroundWindow("C4 · purchase as CUSTOMER 2 (set up the race)",
			"# Auth here is CUSTOMER 2 — buy a pet customer 1 is viewing, then watch\n"+
				"# customer 1 get a human-readable \"no longer available\" error.\n"+
				"mutation PurchasePet($petId: ID!) {\n  purchasePet(petId: $petId) { id name status }\n}\n",
			playgroundCustomer2Auth, "{\n  \"petId\": \"PASTE-A-PET-CUSTOMER-1-WILL-TRY\"\n}"),
	}
}
