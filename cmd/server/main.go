package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"roboticCrewChallenge/internal/adapter/objectstore"
	"roboticCrewChallenge/internal/adapter/postgres"
	"roboticCrewChallenge/internal/adapter/rediscache"
	"roboticCrewChallenge/internal/app/listing"
	"roboticCrewChallenge/internal/app/purchase"
	"roboticCrewChallenge/internal/auth"
	"roboticCrewChallenge/internal/config"
	"roboticCrewChallenge/internal/graph"
	"roboticCrewChallenge/internal/platform/crypto"
	"roboticCrewChallenge/internal/platform/logging"
	"roboticCrewChallenge/internal/server"
)

// Catalog correctness comes from generation-bump invalidation on every write, not
// from this TTL — so it is only a memory-reclamation backstop for stores that go
// quiet, set well above a browsing session to avoid needless refills (and the
// refill latency spike they cause under heavy concurrent load). A write still
// makes its store's listings reachable-stale for zero seconds.
const catalogCacheTTL = 10 * time.Minute

// The catalog read path touches Redis on every request (generation + page lookup),
// so the pool is sized for the concurrent-user target and keeps idle connections
// warm: a cold pool forced to open connections under a traffic burst is the
// dominant source of first-request latency spikes (see docs/PERFORMANCE.md).
const (
	redisPoolSize     = 200
	redisMinIdleConns = 50
)

func main() {
	if err := run(); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := logging.New(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	pictureStore, err := objectstore.New(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOBucket, cfg.MinIOUseSSL)
	if err != nil {
		return err
	}
	if err := pictureStore.EnsureBucket(ctx); err != nil {
		return err
	}

	encryptor, err := crypto.NewEncryptor(cfg.PIIEncryptionKey)
	if err != nil {
		return err
	}
	blindIndex, err := crypto.NewBlindIndex(cfg.PIIEncryptionKey)
	if err != nil {
		return err
	}

	authenticator := auth.NewAuthenticator(
		postgres.NewMerchantRepository(pool, encryptor, blindIndex),
		postgres.NewCustomerRepository(pool, encryptor, blindIndex),
		postgres.NewStoreRepository(pool),
	)

	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		PoolSize:     redisPoolSize,
		MinIdleConns: redisMinIdleConns,
	})
	defer func() { _ = redisClient.Close() }()
	pingCtx, cancelPing := context.WithTimeout(ctx, 2*time.Second)
	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		logger.Warn("redis unreachable at startup; catalog cache degraded to passthrough", "addr", cfg.RedisAddr, "error", err)
	}
	cancelPing()
	catalogCache := rediscache.New(redisClient, encryptor, logger, catalogCacheTTL)

	petRepo := postgres.NewPetRepository(pool, encryptor)
	listingService := listing.NewService(petRepo, pictureStore, catalogCache)
	purchaseService := purchase.NewService(petRepo, catalogCache)
	graphqlHandler := graph.NewHandler(&graph.Resolver{
		Listing:  listingService,
		Purchase: purchaseService,
	}, logger, cfg.GraphQLIntrospection)

	var playgroundHandler http.Handler
	if cfg.GraphQLIntrospection {
		playgroundHandler = graph.NewPlaygroundHandler("/graphql")
	}

	srv := server.New(cfg, logger, authenticator, graphqlHandler, playgroundHandler, pictureStore)
	return srv.Run(ctx)
}
