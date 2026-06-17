package graph

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"roboticCrewChallenge/internal/graph/generated"
	"roboticCrewChallenge/internal/platform/picture"
)

const (
	queryCacheSize    = 1000
	multipartHeadroom = 1 << 20
	// complexityLimit bounds a query's cost. A legitimate full page (first up to
	// the 100 page cap, every node field selected) costs well under this; the
	// scorer multiplies by the raw requested `first` (before the resolver clamps
	// it), so a huge `first` is rejected before any resolver runs — closing the
	// list-amplification attack surface.
	complexityLimit = 4000
)

// NewHandler builds the GraphQL HTTP handler: POST for queries and mutations,
// multipart (size-capped) picture uploads, a single error-translation boundary,
// a panic recovery that never leaks internal detail, and a query complexity
// limit. Introspection stays disabled (gqlgen's default) unless explicitly
// enabled for development.
func NewHandler(resolver *Resolver, logger *slog.Logger, introspection bool) http.Handler {
	srv := handler.New(generated.NewExecutableSchema(executableConfig(resolver)))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{
		MaxUploadSize: picture.MaxPictureBytes + multipartHeadroom,
		MaxMemory:     picture.MaxPictureBytes,
	})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](queryCacheSize))
	srv.Use(extension.FixedComplexityLimit(complexityLimit))
	if introspection {
		srv.Use(extension.Introspection{})
	}
	srv.SetErrorPresenter(PresentError)
	srv.SetRecoverFunc(func(ctx context.Context, err any) error {
		logger.ErrorContext(ctx, "graphql panic recovered", "panic", err)
		return gqlerror.Errorf("internal server error")
	})

	return srv
}

// NewPlaygroundHandler serves the GraphiQL page that a developer opens in a
// browser to explore the schema and run operations against the GraphQL endpoint
// at the given path. The page
// itself is unauthenticated so it can load; the queries it sends carry whatever
// Authorization header the developer sets. It relies on introspection for schema
// docs and autocompletion, so the server only mounts it when introspection is on.
func NewPlaygroundHandler(endpoint string) http.Handler {
	return playground.Handler("Pet Store API", endpoint)
}

// executableConfig wires resolvers plus per-field complexity. The paginated list
// queries cost `first * childComplexity`, using the requested page size (before
// the resolver clamps it) so an abusive `first` is expensive.
func executableConfig(resolver *Resolver) generated.Config {
	cfg := generated.Config{Resolvers: resolver}
	cfg.Complexity.Query.AvailablePets = func(childComplexity int, _ string, first *int, _ *string) int {
		return complexityFirst(first) * childComplexity
	}
	cfg.Complexity.Query.UnsoldPets = func(childComplexity int, first *int, _ *string) int {
		return complexityFirst(first) * childComplexity
	}
	cfg.Complexity.Query.SoldPets = func(childComplexity int, _ time.Time, _ time.Time, first *int, _ *string) int {
		return complexityFirst(first) * childComplexity
	}
	return cfg
}

// complexityFirst mirrors the resolver's nil/non-positive default but, unlike
// clampFirst, deliberately does NOT cap at maxPageSize — the scorer must see the
// real requested size so an abusive `first` is charged its true cost.
func complexityFirst(first *int) int {
	if first == nil || *first <= 0 {
		return defaultPageSize
	}
	return *first
}
