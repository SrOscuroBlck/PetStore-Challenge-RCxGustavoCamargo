package graph

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"roboticCrewChallenge/internal/auth"
	"roboticCrewChallenge/internal/domain"
)

// PresentError is the single boundary that turns the typed errors bubbling up
// from the domain and app layers into GraphQL errors carrying a stable,
// machine-readable code. A recognised domain/auth error gets its code and a safe
// message; a gqlgen validation/parse error (the client's fault) keeps its own
// descriptive message; anything else collapses to INTERNAL so server-side detail
// never reaches a client.
func PresentError(ctx context.Context, err error) *gqlerror.Error {
	gqlErr := graphql.DefaultErrorPresenter(ctx, err)
	if gqlErr.Extensions == nil {
		gqlErr.Extensions = map[string]any{}
	}

	if code, message, ok := classify(err); ok {
		gqlErr.Extensions["code"] = code
		gqlErr.Message = message
		return gqlErr
	}

	// gqlgen raises validation/parse failures as a *gqlerror.Error with a Rule
	// (e.g. a query selecting a non-existent field). These are client errors, so
	// keep gqlgen's own message and tag a stable code rather than masking them.
	var validationErr *gqlerror.Error
	if errors.As(err, &validationErr) && validationErr.Rule != "" {
		gqlErr.Extensions["code"] = "GRAPHQL_VALIDATION_FAILED"
		return gqlErr
	}

	// gqlgen extensions (e.g. the complexity limit) tag the error with a stable
	// code before it reaches here. Preserve it rather than masking it.
	if code, ok := gqlErr.Extensions["code"].(string); ok && code != "" {
		return gqlErr
	}

	gqlErr.Extensions["code"] = "INTERNAL"
	gqlErr.Message = "internal server error"
	return gqlErr
}

func classify(err error) (code string, message string, ok bool) {
	var validation *domain.ValidationError
	var unavailablePets *domain.UnavailablePetsError
	switch {
	case errors.As(err, &validation):
		return "VALIDATION", validation.Error(), true
	case errors.As(err, &unavailablePets):
		return "UNAVAILABLE", unavailablePets.Error(), true
	case errors.Is(err, auth.ErrUnauthenticated):
		return "UNAUTHENTICATED", "authentication is required", true
	case errors.Is(err, auth.ErrForbidden):
		return "FORBIDDEN", "this operation is not permitted for your role", true
	case errors.Is(err, domain.ErrPetNotFound), errors.Is(err, domain.ErrStoreNotFound):
		return "NOT_FOUND", "pet not found", true
	case errors.Is(err, domain.ErrPetNotRemovable):
		return "CONFLICT", "pet cannot be removed because it is no longer available", true
	case errors.Is(err, domain.ErrPetUnavailable):
		return "UNAVAILABLE", "pet is no longer available", true
	case errors.Is(err, domain.ErrUnsupportedPictureType):
		return "UNSUPPORTED_MEDIA_TYPE", "picture must be a JPEG, PNG, or WebP image", true
	case errors.Is(err, domain.ErrPictureTooLarge):
		return "PAYLOAD_TOO_LARGE", "picture exceeds the maximum allowed size", true
	default:
		return "", "", false
	}
}
