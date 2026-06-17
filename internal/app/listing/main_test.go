package listing_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"roboticCrewChallenge/internal/adapter/postgres"
	"roboticCrewChallenge/internal/adapter/rediscache"
	"roboticCrewChallenge/internal/app/listing"
	"roboticCrewChallenge/internal/testsupport"
)

const testCacheTTL = 60 * time.Second

var (
	harness    *testsupport.Harness
	testLogger = slog.New(slog.NewTextHandler(io.Discard, nil))
	pngBytes   = testsupport.PNGBytes()
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
	h, cleanup, ok, err := testsupport.Start(ctx, testsupport.Options{SchemaPath: schema, WithPostgres: true, WithMinIO: true, WithRedis: true})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if !ok {
		return m.Run()
	}
	defer cleanup()
	harness = h
	return m.Run()
}

func requireInfra(t *testing.T) {
	t.Helper()
	if harness == nil {
		t.Skip("postgres/minio/redis containers unavailable")
	}
}

func newService() *listing.Service {
	return listing.NewService(
		postgres.NewPetRepository(harness.Pool, harness.Enc),
		harness.PictureStore,
		rediscache.New(harness.RedisClient, harness.Enc, testLogger, testCacheTTL),
	)
}

func seedStore(t *testing.T) uuid.UUID    { return harness.SeedStore(t) }
func seedCustomer(t *testing.T) uuid.UUID { return harness.SeedCustomer(t) }
