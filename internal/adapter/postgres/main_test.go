package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/platform/crypto"
)

var (
	testPool  *pgxpool.Pool
	testEnc   *crypto.Encryptor
	testBlind *crypto.BlindIndex
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()

	schema, err := filepath.Abs(filepath.Join("..", "..", "..", "db", "schema", "schema.sql"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "resolve schema path:", err)
		return 1
	}

	container, err := tcpostgres.Run(ctx, "postgres:16",
		tcpostgres.WithDatabase("petstore_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.WithInitScripts(schema),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		// In CI a failure to start the container is a real failure (a false-green run would
		// hide untested repositories). Locally, where Docker may be absent, skip instead.
		if os.Getenv("CI") != "" {
			fmt.Fprintln(os.Stderr, "postgres integration tests failed to start in CI:", err)
			return 1
		}
		fmt.Fprintln(os.Stderr, "skipping postgres integration tests (Docker unavailable):", err)
		return 0
	}
	defer func() { _ = container.Terminate(ctx) }()

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
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

	return m.Run()
}

// seedStore creates a merchant and its store, returning the store id for pet tests.
func seedStore(t *testing.T) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	merchantID := uuid.New()
	merchant, err := domain.NewMerchant(merchantID, fmt.Sprintf("m-%s@example.com", merchantID), "hashed", time.Now())
	if err != nil {
		t.Fatalf("build merchant: %v", err)
	}
	if err := NewMerchantRepository(testPool, testEnc, testBlind).Create(ctx, merchant); err != nil {
		t.Fatalf("create merchant: %v", err)
	}
	storeID := uuid.New()
	store, err := domain.NewStore(storeID, merchantID, "Pluto's Pets", time.Now())
	if err != nil {
		t.Fatalf("build store: %v", err)
	}
	if err := NewStoreRepository(testPool).Create(ctx, store); err != nil {
		t.Fatalf("create store: %v", err)
	}
	return storeID
}

func newPet(t *testing.T, storeID uuid.UUID) domain.Pet {
	t.Helper()
	pet, err := domain.NewPet(domain.NewPetParams{
		ID:               uuid.New(),
		StoreID:          storeID,
		Name:             "Pluto",
		Species:          "DOG",
		AgeYears:         3,
		Description:      "Friendly",
		BreederName:      "Jane Doe",
		BreederEmail:     "jane@example.com",
		PictureObjectKey: "pets/pluto.jpg",
		CreatedAt:        time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("build pet: %v", err)
	}
	return pet
}
