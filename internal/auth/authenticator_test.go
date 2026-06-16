package auth_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"roboticCrewChallenge/internal/adapter/postgres"
	"roboticCrewChallenge/internal/auth"
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

	schema, err := filepath.Abs(filepath.Join("..", "..", "db", "schema", "schema.sql"))
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
		if os.Getenv("CI") != "" {
			fmt.Fprintln(os.Stderr, "auth integration tests failed to start in CI:", err)
			return 1
		}
		fmt.Fprintln(os.Stderr, "skipping auth integration tests (Docker unavailable):", err)
		return m.Run()
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

func newAuthenticator() *auth.Authenticator {
	return auth.NewAuthenticator(
		postgres.NewMerchantRepository(testPool, testEnc, testBlind),
		postgres.NewCustomerRepository(testPool, testEnc, testBlind),
		postgres.NewStoreRepository(testPool),
	)
}

const testPassword = "correct horse battery staple"

// seedMerchant creates a merchant with a store and returns the merchant id,
// store id, and login email.
func seedMerchant(t *testing.T) (uuid.UUID, uuid.UUID, string) {
	t.Helper()
	ctx := context.Background()
	merchantID := uuid.New()
	email := fmt.Sprintf("merchant-%s@example.com", merchantID)
	hash, err := crypto.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	merchant, err := domain.NewMerchant(merchantID, email, hash, time.Now().UTC())
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
	return merchantID, storeID, email
}

func seedCustomer(t *testing.T) (uuid.UUID, string) {
	t.Helper()
	ctx := context.Background()
	customerID := uuid.New()
	email := fmt.Sprintf("customer-%s@example.com", customerID)
	hash, err := crypto.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	customer, err := domain.NewCustomer(customerID, email, hash, time.Now().UTC())
	if err != nil {
		t.Fatalf("build customer: %v", err)
	}
	if err := postgres.NewCustomerRepository(testPool, testEnc, testBlind).Create(ctx, customer); err != nil {
		t.Fatalf("create customer: %v", err)
	}
	return customerID, email
}

func seedPet(t *testing.T, storeID uuid.UUID) uuid.UUID {
	t.Helper()
	petID := uuid.New()
	pet, err := domain.NewPet(domain.NewPetParams{
		ID:               petID,
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
	if err := postgres.NewPetRepository(testPool, testEnc).Create(context.Background(), pet); err != nil {
		t.Fatalf("create pet: %v", err)
	}
	return petID
}

// AC#1: a request without valid credentials is rejected, and an unknown email is
// indistinguishable from a wrong password.
func TestAuthenticate_RejectsInvalidCredentials(t *testing.T) {
	requireDB(t)
	ctx := context.Background()
	_, _, email := seedMerchant(t)
	authenticator := newAuthenticator()

	if _, err := authenticator.Authenticate(ctx, "nobody@example.com", testPassword); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("unknown email: expected ErrInvalidCredentials, got %v", err)
	}
	if _, err := authenticator.Authenticate(ctx, email, "wrong password"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("wrong password: expected ErrInvalidCredentials, got %v", err)
	}
}

// AC#2: roles are resolved from the credentials and enforced by the guards.
func TestAuthenticate_ResolvesRolesAndGuardsSeparateThem(t *testing.T) {
	requireDB(t)
	ctx := context.Background()
	_, storeID, merchantEmail := seedMerchant(t)
	_, customerEmail := seedCustomer(t)
	authenticator := newAuthenticator()

	merchant, err := authenticator.Authenticate(ctx, merchantEmail, testPassword)
	if err != nil {
		t.Fatalf("merchant authenticate: %v", err)
	}
	if merchant.Role != auth.RoleMerchant || merchant.StoreID != storeID {
		t.Fatalf("merchant identity = %+v, want role merchant and store %s", merchant, storeID)
	}

	customer, err := authenticator.Authenticate(ctx, customerEmail, testPassword)
	if err != nil {
		t.Fatalf("customer authenticate: %v", err)
	}
	if customer.Role != auth.RoleCustomer || customer.StoreID != uuid.Nil {
		t.Fatalf("customer identity = %+v, want role customer and no store", customer)
	}

	merchantCtx := auth.WithIdentity(ctx, merchant)
	customerCtx := auth.WithIdentity(ctx, customer)
	if _, err := auth.RequireCustomer(merchantCtx); !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("merchant on a customer guard: expected ErrForbidden, got %v", err)
	}
	if _, err := auth.RequireMerchant(customerCtx); !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("customer on a merchant guard: expected ErrForbidden, got %v", err)
	}
}

// AC#3: a merchant authenticated for store B cannot reach store A's pet — the
// store-scoped lookup returns not-found, never the other store's row.
func TestStoreIsolation_CrossStorePetIsNotFound(t *testing.T) {
	requireDB(t)
	ctx := context.Background()
	_, storeA, _ := seedMerchant(t)
	petA := seedPet(t, storeA)
	_, _, emailB := seedMerchant(t)
	authenticator := newAuthenticator()

	identityB, err := authenticator.Authenticate(ctx, emailB, testPassword)
	if err != nil {
		t.Fatalf("merchant B authenticate: %v", err)
	}

	petRepo := postgres.NewPetRepository(testPool, testEnc)
	if _, err := petRepo.GetByID(ctx, identityB.StoreID, petA); !errors.Is(err, domain.ErrPetNotFound) {
		t.Fatalf("cross-store GetByID: expected ErrPetNotFound, got %v", err)
	}
	if _, err := petRepo.Remove(ctx, identityB.StoreID, petA); !errors.Is(err, domain.ErrPetNotFound) {
		t.Fatalf("cross-store Remove: expected ErrPetNotFound, got %v", err)
	}
}

// AC#4: passwords are stored only as bcrypt hashes, never the plaintext.
func TestPasswordsStoredAsBcrypt(t *testing.T) {
	requireDB(t)
	ctx := context.Background()
	merchantID, _, _ := seedMerchant(t)

	merchant, err := postgres.NewMerchantRepository(testPool, testEnc, testBlind).GetByID(ctx, merchantID)
	if err != nil {
		t.Fatalf("get merchant: %v", err)
	}
	if !strings.HasPrefix(merchant.PasswordHash, "$2") {
		t.Fatalf("expected a bcrypt hash, got %q", merchant.PasswordHash)
	}
	if strings.Contains(merchant.PasswordHash, testPassword) {
		t.Fatal("stored hash must not contain the plaintext password")
	}
}

func requireDB(t *testing.T) {
	t.Helper()
	if testPool == nil {
		t.Skip("postgres container unavailable")
	}
}
