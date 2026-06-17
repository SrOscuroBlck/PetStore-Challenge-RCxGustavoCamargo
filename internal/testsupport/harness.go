// Package testsupport brings up the real backing services (Postgres, MinIO,
// Redis) for integration tests via testcontainers, sharing one set of helpers
// across the packages that need them. It centralises the CI-fails / local-skips
// contract: in CI a container that will not start is a hard failure (a
// false-green run would hide untested code), while locally — where Docker may be
// absent — the suite skips instead.
package testsupport

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"roboticCrewChallenge/internal/adapter/objectstore"
	"roboticCrewChallenge/internal/platform/crypto"
)

// Options selects which backing services a package needs. SchemaPath is the
// absolute path to db/schema/schema.sql, resolved by the caller (which knows its
// own directory depth); it is required only when WithPostgres is set.
type Options struct {
	SchemaPath   string
	WithPostgres bool
	WithMinIO    bool
	WithRedis    bool
}

// Harness holds the live clients for a test package. Fields are populated only
// when the corresponding Options flag is set.
type Harness struct {
	Pool         *pgxpool.Pool
	Enc          *crypto.Encryptor
	Blind        *crypto.BlindIndex
	RedisClient  *redis.Client
	PictureStore *objectstore.PictureStore
}

// Start brings up the requested containers. ok is false when the infrastructure
// is unavailable locally (the package should skip); a non-nil error means a hard
// failure (in CI, or a genuine setup error) and the caller should fail the run.
// The returned cleanup tears down everything Start created.
func Start(ctx context.Context, opts Options) (h *Harness, cleanup func(), ok bool, err error) {
	var teardowns []func()
	cleanup = func() {
		for i := len(teardowns) - 1; i >= 0; i-- {
			teardowns[i]()
		}
	}

	enc, blind, err := newCrypto()
	if err != nil {
		cleanup()
		return nil, nil, false, err
	}
	result := &Harness{Enc: enc, Blind: blind}

	if opts.WithPostgres {
		pg, err := tcpostgres.Run(ctx, "postgres:16",
			tcpostgres.WithDatabase("petstore_test"),
			tcpostgres.WithUsername("test"),
			tcpostgres.WithPassword("test"),
			tcpostgres.WithInitScripts(opts.SchemaPath),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).WithStartupTimeout(60*time.Second),
			),
		)
		if err != nil {
			return skipOrFail(cleanup, "postgres", err)
		}
		teardowns = append(teardowns, func() { _ = pg.Terminate(ctx) })

		dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			cleanup()
			return nil, nil, false, fmt.Errorf("postgres connection string: %w", err)
		}
		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			cleanup()
			return nil, nil, false, fmt.Errorf("create pool: %w", err)
		}
		teardowns = append(teardowns, pool.Close)
		result.Pool = pool
	}

	if opts.WithMinIO {
		store, terminate, err := startMinIO(ctx)
		if isContainerStartFailure(err) {
			return skipOrFail(cleanup, "minio", err)
		}
		if err != nil {
			cleanup()
			return nil, nil, false, err
		}
		teardowns = append(teardowns, terminate)
		result.PictureStore = store
	}

	if opts.WithRedis {
		client, terminate, err := startRedis(ctx)
		if isContainerStartFailure(err) {
			return skipOrFail(cleanup, "redis", err)
		}
		if err != nil {
			cleanup()
			return nil, nil, false, err
		}
		teardowns = append(teardowns, terminate)
		result.RedisClient = client
	}

	return result, cleanup, true, nil
}

func skipOrFail(cleanup func(), service string, err error) (*Harness, func(), bool, error) {
	cleanup()
	if os.Getenv("CI") != "" {
		return nil, nil, false, fmt.Errorf("%s integration container failed to start in CI: %w", service, err)
	}
	fmt.Fprintln(os.Stderr, "skipping integration tests (Docker unavailable):", err)
	return nil, nil, false, nil
}

func newCrypto() (*crypto.Encryptor, *crypto.BlindIndex, error) {
	key := make([]byte, crypto.KeySize)
	for i := range key {
		key[i] = byte(i + 1)
	}
	enc, err := crypto.NewEncryptor(key)
	if err != nil {
		return nil, nil, fmt.Errorf("test encryptor: %w", err)
	}
	blind, err := crypto.NewBlindIndex(key)
	if err != nil {
		return nil, nil, fmt.Errorf("test blind index: %w", err)
	}
	return enc, blind, nil
}

type containerStartError struct{ err error }

func (e containerStartError) Error() string { return e.err.Error() }
func (e containerStartError) Unwrap() error { return e.err }

func isContainerStartFailure(err error) bool {
	var startErr containerStartError
	return errors.As(err, &startErr)
}

func startMinIO(ctx context.Context) (*objectstore.PictureStore, func(), error) {
	container, err := tcminio.Run(ctx, "minio/minio:RELEASE.2024-01-16T16-07-38Z")
	if err != nil {
		return nil, nil, containerStartError{err}
	}
	terminate := func() { _ = container.Terminate(ctx) }
	endpoint, err := container.ConnectionString(ctx)
	if err != nil {
		terminate()
		return nil, nil, fmt.Errorf("minio connection string: %w", err)
	}
	store, err := objectstore.New(endpoint, container.Username, container.Password, "pet-pictures-test", false)
	if err != nil {
		terminate()
		return nil, nil, fmt.Errorf("new picture store: %w", err)
	}
	if err := store.EnsureBucket(ctx); err != nil {
		terminate()
		return nil, nil, fmt.Errorf("ensure bucket: %w", err)
	}
	return store, terminate, nil
}

func startRedis(ctx context.Context) (*redis.Client, func(), error) {
	container, err := tcredis.Run(ctx, "redis:7")
	if err != nil {
		return nil, nil, containerStartError{err}
	}
	terminate := func() { _ = container.Terminate(ctx) }
	endpoint, err := container.ConnectionString(ctx)
	if err != nil {
		terminate()
		return nil, nil, fmt.Errorf("redis connection string: %w", err)
	}
	opts, err := redis.ParseURL(endpoint)
	if err != nil {
		terminate()
		return nil, nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(opts)
	return client, func() { _ = client.Close(); terminate() }, nil
}
