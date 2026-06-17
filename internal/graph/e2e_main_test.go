package graph

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"roboticCrewChallenge/internal/adapter/postgres"
	"roboticCrewChallenge/internal/adapter/rediscache"
	"roboticCrewChallenge/internal/app/listing"
	"roboticCrewChallenge/internal/app/purchase"
	"roboticCrewChallenge/internal/auth"
	"roboticCrewChallenge/internal/testsupport"
)

var (
	harness     *testsupport.Harness
	testHandler http.Handler
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()
	schema, err := filepath.Abs(filepath.Join("..", "..", "db", "schema", "schema.sql"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "resolve schema path:", err)
		return 1
	}
	h, cleanup, ok, err := testsupport.Start(ctx, testsupport.Options{SchemaPath: schema, WithPostgres: true, WithMinIO: true})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if !ok {
		return m.Run()
	}
	defer cleanup()
	harness = h

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	petRepo := postgres.NewPetRepository(h.Pool, h.Enc)
	listingService := listing.NewService(petRepo, h.PictureStore, rediscache.NoOp{})
	purchaseService := purchase.NewService(petRepo, rediscache.NoOp{})
	testHandler = NewHandler(&Resolver{
		Listing:  listingService,
		Purchase: purchaseService,
	}, logger, false)

	return m.Run()
}

func requireInfra(t *testing.T) {
	t.Helper()
	if testHandler == nil {
		t.Skip("postgres/minio containers unavailable")
	}
}

// handlerAs wraps the GraphQL handler so each request carries the given identity
// in context, standing in for what the Basic-auth middleware injects in production.
func handlerAs(identity *auth.Identity) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if identity != nil {
			r = r.WithContext(auth.WithIdentity(r.Context(), *identity))
		}
		testHandler.ServeHTTP(w, r)
	})
}

func merchantIdentity(storeID uuid.UUID) auth.Identity {
	return auth.Identity{Subject: uuid.New(), Role: auth.RoleMerchant, StoreID: storeID}
}

// customerIdentity seeds a real customer (its id is an FK target for sold_by) and
// returns the matching identity.
func customerIdentity(t *testing.T) auth.Identity {
	return auth.Identity{Subject: seedCustomer(t), Role: auth.RoleCustomer}
}

func seedStore(t *testing.T) uuid.UUID    { return harness.SeedStore(t) }
func seedCustomer(t *testing.T) uuid.UUID { return harness.SeedCustomer(t) }

func writePNG(t *testing.T) *os.File {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "pet-*.png")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	if _, err := f.Write(testsupport.PNGBytes()); err != nil {
		t.Fatalf("write png: %v", err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("seek: %v", err)
	}
	return f
}
