package graph

import (
	"errors"
	"fmt"
	"testing"

	"roboticCrewChallenge/internal/auth"
	"roboticCrewChallenge/internal/domain"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		name string
		err  error
		code string
	}{
		{"validation", &domain.ValidationError{Field: "name", Msg: "is required"}, "VALIDATION"},
		{"unauthenticated", auth.ErrUnauthenticated, "UNAUTHENTICATED"},
		{"forbidden", auth.ErrForbidden, "FORBIDDEN"},
		{"pet not found", domain.ErrPetNotFound, "NOT_FOUND"},
		{"store not found", domain.ErrStoreNotFound, "NOT_FOUND"},
		{"not removable", domain.ErrPetNotRemovable, "CONFLICT"},
		{"unavailable", domain.ErrPetUnavailable, "UNAVAILABLE"},
		{"bad picture type", domain.ErrUnsupportedPictureType, "UNSUPPORTED_MEDIA_TYPE"},
		{"picture too large", domain.ErrPictureTooLarge, "PAYLOAD_TOO_LARGE"},
		{"unknown", errors.New("boom"), "INTERNAL"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, message := classify(tc.err)
			if code != tc.code {
				t.Fatalf("code = %q, want %q", code, tc.code)
			}
			if message == "" {
				t.Fatal("message must not be empty")
			}
			wrappedCode, _ := classify(fmt.Errorf("layer: %w", tc.err))
			if wrappedCode != tc.code {
				t.Fatalf("wrapped code = %q, want %q (errors.Is/As must unwrap)", wrappedCode, tc.code)
			}
		})
	}
}

func TestClassify_UnknownLeaksNothing(t *testing.T) {
	_, message := classify(errors.New("sensitive internal detail: dsn=postgres://secret"))
	if message != "internal server error" {
		t.Fatalf("unknown error message = %q, must be the generic message", message)
	}
}
