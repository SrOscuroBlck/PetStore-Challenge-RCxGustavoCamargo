package graph

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"roboticCrewChallenge/internal/graph/generated"
	"roboticCrewChallenge/internal/platform/picture"
)

const (
	queryCacheSize    = 1000
	multipartHeadroom = 1 << 20
)

// NewHandler builds the GraphQL HTTP handler: POST for queries and mutations,
// multipart for picture uploads, a single error-translation boundary, and a
// panic recovery that never leaks internal detail. Introspection and complexity
// limits are intentionally left to the hardening issue.
func NewHandler(resolver *Resolver, logger *slog.Logger) http.Handler {
	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{
		MaxUploadSize: picture.MaxPictureBytes + multipartHeadroom,
		MaxMemory:     picture.MaxPictureBytes,
	})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](queryCacheSize))
	srv.SetErrorPresenter(PresentError)
	srv.SetRecoverFunc(func(ctx context.Context, err any) error {
		logger.ErrorContext(ctx, "graphql panic recovered", "panic", err)
		return gqlerror.Errorf("internal server error")
	})

	return srv
}
