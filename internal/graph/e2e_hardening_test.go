package graph

import (
	"testing"

	"github.com/99designs/gqlgen/client"
)

func TestE2E_Introspection_DisabledByDefault(t *testing.T) {
	requireInfra(t)
	c := client.New(handlerAs(nil))

	resp, err := c.RawPost(`{ __schema { queryType { name } } }`)
	if err != nil {
		t.Fatalf("raw post: %v", err)
	}
	if len(resp.Errors) == 0 {
		t.Fatal("introspection must be rejected when disabled")
	}
}

func TestE2E_AvailablePets_ComplexityLimitRejectsHugeFirst(t *testing.T) {
	requireInfra(t)
	storeID := seedStore(t)
	c := client.New(handlerAs(ptr(customerIdentity(t))))

	const query = `query($s: ID!) {
  availablePets(storeId: $s, first: 100000) {
    edges { node { id name species ageYears description pictureUrl status createdAt soldAt } cursor }
    pageInfo { hasNextPage endCursor }
  }
}`
	if code := errorCode(t, c, query, client.Var("s", storeID.String())); code != "COMPLEXITY_LIMIT_EXCEEDED" {
		t.Fatalf("error code = %q, want COMPLEXITY_LIMIT_EXCEEDED", code)
	}
}
