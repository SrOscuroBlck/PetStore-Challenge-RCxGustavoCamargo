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
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"roboticCrewChallenge/internal/adapter/objectstore"
	"roboticCrewChallenge/internal/adapter/postgres"
	"roboticCrewChallenge/internal/app/listing"
	"roboticCrewChallenge/internal/auth"
	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/platform/crypto"
)

var (
	testPool    *pgxpool.Pool
	testEnc     *crypto.Encryptor
	testBlind   *crypto.BlindIndex
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

	pg, err := tcpostgres.Run(ctx, "postgres:16",
		tcpostgres.WithDatabase("petstore_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.WithInitScripts(schema),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		if os.Getenv("CI") != "" {
			fmt.Fprintln(os.Stderr, "graph e2e tests failed to start postgres in CI:", err)
			return 1
		}
		fmt.Fprintln(os.Stderr, "skipping graph e2e tests (Docker unavailable):", err)
		return m.Run()
	}
	defer func() { _ = pg.Terminate(ctx) }()

	mn, err := tcminio.Run(ctx, "minio/minio:RELEASE.2024-01-16T16-07-38Z")
	if err != nil {
		if os.Getenv("CI") != "" {
			fmt.Fprintln(os.Stderr, "graph e2e tests failed to start minio in CI:", err)
			return 1
		}
		fmt.Fprintln(os.Stderr, "skipping graph e2e tests (Docker unavailable):", err)
		return m.Run()
	}
	defer func() { _ = mn.Terminate(ctx) }()

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintln(os.Stderr, "connection string:", err)
		return 1
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create pool:", err)
		return 1
	}
	defer pool.Close()
	testPool = pool

	key := make([]byte, crypto.KeySize)
	for i := range key {
		key[i] = byte(i + 1)
	}
	if testEnc, err = crypto.NewEncryptor(key); err != nil {
		fmt.Fprintln(os.Stderr, "test encryptor:", err)
		return 1
	}
	if testBlind, err = crypto.NewBlindIndex(key); err != nil {
		fmt.Fprintln(os.Stderr, "test blind index:", err)
		return 1
	}

	endpoint, err := mn.ConnectionString(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "minio connection string:", err)
		return 1
	}
	pictures, err := objectstore.New(endpoint, mn.Username, mn.Password, "pet-pictures-test", false)
	if err != nil {
		fmt.Fprintln(os.Stderr, "new picture store:", err)
		return 1
	}
	if err := pictures.EnsureBucket(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "ensure bucket:", err)
		return 1
	}

	service := listing.NewService(postgres.NewPetRepository(pool, testEnc), pictures)
	testHandler = NewHandler(&Resolver{Listing: service, PictureStore: pictures}, slog.New(slog.NewTextHandler(io.Discard, nil)))

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

func seedStore(t *testing.T) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	merchantID := uuid.New()
	merchant, err := domain.NewMerchant(merchantID, fmt.Sprintf("m-%s@example.com", merchantID), "hash", time.Now().UTC())
	if err != nil {
		t.Fatalf("build merchant: %v", err)
	}
	if err := postgres.NewMerchantRepository(testPool, testEnc, testBlind).Create(ctx, merchant); err != nil {
		t.Fatalf("create merchant: %v", err)
	}
	storeID := uuid.New()
	store, err := domain.NewStore(storeID, merchantID, "Pluto's Pets", time.Now().UTC())
	if err != nil {
		t.Fatalf("build store: %v", err)
	}
	if err := postgres.NewStoreRepository(testPool).Create(ctx, store); err != nil {
		t.Fatalf("create store: %v", err)
	}
	return storeID
}

func seedCustomer(t *testing.T) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	customerID := uuid.New()
	customer, err := domain.NewCustomer(customerID, fmt.Sprintf("c-%s@example.com", customerID), "hash", time.Now().UTC())
	if err != nil {
		t.Fatalf("build customer: %v", err)
	}
	if err := postgres.NewCustomerRepository(testPool, testEnc, testBlind).Create(ctx, customer); err != nil {
		t.Fatalf("create customer: %v", err)
	}
	return customerID
}

func writePNG(t *testing.T) *os.File {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "pet-*.png")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	if _, err := f.Write(append([]byte("\x89PNG\r\n\x1a\n"), []byte("sniffable-png-data")...)); err != nil {
		t.Fatalf("write png: %v", err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("seek: %v", err)
	}
	return f
}
