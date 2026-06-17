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
// machine-readable code. Unrecognised errors collapse to INTERNAL so internal
// detail never reaches a client.
func PresentError(ctx context.Context, err error) *gqlerror.Error {
	gqlErr := graphql.DefaultErrorPresenter(ctx, err)
	code, message := classify(err)
	if gqlErr.Extensions == nil {
		gqlErr.Extensions = map[string]any{}
	}
	gqlErr.Extensions["code"] = code
	gqlErr.Message = message
	return gqlErr
}

func classify(err error) (code string, message string) {
	var validation *domain.ValidationError
	switch {
	case errors.As(err, &validation):
		return "VALIDATION", validation.Error()
	case errors.Is(err, auth.ErrUnauthenticated):
		return "UNAUTHENTICATED", "authentication is required"
	case errors.Is(err, auth.ErrForbidden):
		return "FORBIDDEN", "this operation is not permitted for your role"
	case errors.Is(err, domain.ErrPetNotFound), errors.Is(err, domain.ErrStoreNotFound):
		return "NOT_FOUND", "pet not found"
	case errors.Is(err, domain.ErrPetNotRemovable):
		return "CONFLICT", "pet cannot be removed because it is no longer available"
	case errors.Is(err, domain.ErrPetUnavailable):
		return "UNAVAILABLE", "pet is no longer available"
	case errors.Is(err, domain.ErrUnsupportedPictureType):
		return "UNSUPPORTED_MEDIA_TYPE", "picture must be a JPEG, PNG, or WebP image"
	case errors.Is(err, domain.ErrPictureTooLarge):
		return "PAYLOAD_TOO_LARGE", "picture exceeds the maximum allowed size"
	default:
		return "INTERNAL", "internal server error"
	}
}
