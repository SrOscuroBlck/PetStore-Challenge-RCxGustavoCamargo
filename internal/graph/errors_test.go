package graph

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/vektah/gqlparser/v2/gqlerror"

	"roboticCrewChallenge/internal/auth"
	"roboticCrewChallenge/internal/domain"
)

func TestClassify_RecognisedErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
		code string
	}{
		{"validation", &domain.ValidationError{Field: "name", Msg: "is required"}, "VALIDATION"},
		{"unavailable pets", &domain.UnavailablePetsError{Pets: []domain.UnavailablePet{{Name: "Pluto"}}}, "UNAVAILABLE"},
		{"unauthenticated", auth.ErrUnauthenticated, "UNAUTHENTICATED"},
		{"forbidden", auth.ErrForbidden, "FORBIDDEN"},
		{"pet not found", domain.ErrPetNotFound, "NOT_FOUND"},
		{"store not found", domain.ErrStoreNotFound, "NOT_FOUND"},
		{"not removable", domain.ErrPetNotRemovable, "CONFLICT"},
		{"unavailable", domain.ErrPetUnavailable, "UNAVAILABLE"},
		{"bad picture type", domain.ErrUnsupportedPictureType, "UNSUPPORTED_MEDIA_TYPE"},
		{"picture too large", domain.ErrPictureTooLarge, "PAYLOAD_TOO_LARGE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, message, ok := classify(tc.err)
			if !ok || code != tc.code {
				t.Fatalf("classify = (%q, ok=%v), want code %q", code, ok, tc.code)
			}
			if message == "" {
				t.Fatal("message must not be empty")
			}
			wrappedCode, _, wrappedOK := classify(fmt.Errorf("layer: %w", tc.err))
			if !wrappedOK || wrappedCode != tc.code {
				t.Fatalf("wrapped classify = (%q, ok=%v), want %q (errors.Is/As must unwrap)", wrappedCode, wrappedOK, tc.code)
			}
		})
	}
}

func TestClassify_UnknownIsNotRecognised(t *testing.T) {
	if _, _, ok := classify(errors.New("boom")); ok {
		t.Fatal("an unknown error must not be recognised by classify")
	}
}

func TestPresentError_UnknownLeaksNothing(t *testing.T) {
	gqlErr := PresentError(context.Background(), errors.New("sensitive internal detail: dsn=postgres://secret"))
	if gqlErr.Message != "internal server error" {
		t.Fatalf("unknown error message = %q, must be the generic message", gqlErr.Message)
	}
	if gqlErr.Extensions["code"] != "INTERNAL" {
		t.Fatalf("unknown error code = %v, want INTERNAL", gqlErr.Extensions["code"])
	}
}

func TestPresentError_PreservesValidationError(t *testing.T) {
	validation := &gqlerror.Error{Message: `Cannot query field "breederName" on type "PublicPet".`, Rule: "FieldsOnCorrectType"}
	gqlErr := PresentError(context.Background(), validation)
	if gqlErr.Extensions["code"] != "GRAPHQL_VALIDATION_FAILED" {
		t.Fatalf("validation code = %v, want GRAPHQL_VALIDATION_FAILED", gqlErr.Extensions["code"])
	}
	if gqlErr.Message != validation.Message {
		t.Fatalf("validation message was clobbered: got %q", gqlErr.Message)
	}
}

func TestPresentError_PreservesPreTaggedCode(t *testing.T) {
	// gqlgen's complexity-limit extension tags the error with a stable code
	// before it reaches the presenter; it must not be masked as INTERNAL.
	tagged := &gqlerror.Error{
		Message:    "operation has complexity 9000, which exceeds the limit of 4000",
		Extensions: map[string]any{"code": "COMPLEXITY_LIMIT_EXCEEDED"},
	}
	gqlErr := PresentError(context.Background(), tagged)
	if gqlErr.Extensions["code"] != "COMPLEXITY_LIMIT_EXCEEDED" {
		t.Fatalf("pre-tagged code = %v, want COMPLEXITY_LIMIT_EXCEEDED", gqlErr.Extensions["code"])
	}
	if gqlErr.Message != tagged.Message {
		t.Fatalf("pre-tagged message was clobbered: got %q", gqlErr.Message)
	}
}

func TestPresentError_RecognisedErrorGetsCode(t *testing.T) {
	gqlErr := PresentError(context.Background(), domain.ErrPetUnavailable)
	if gqlErr.Extensions["code"] != "UNAVAILABLE" {
		t.Fatalf("code = %v, want UNAVAILABLE", gqlErr.Extensions["code"])
	}
}
