package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

type fakeAuthenticator struct {
	identity Identity
	err      error
}

func (f fakeAuthenticator) Authenticate(_ context.Context, _, _ string) (Identity, error) {
	return f.identity, f.err
}

func basicHeader(email, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(email+":"+password))
}

func serve(t *testing.T, auth authenticate, header string) (*httptest.ResponseRecorder, *Identity) {
	t.Helper()
	var seen *Identity
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, ok := IdentityFromContext(r.Context()); ok {
			seen = &id
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	rec := httptest.NewRecorder()
	BasicAuth(auth)(next).ServeHTTP(rec, req)
	return rec, seen
}

func TestBasicAuth_MissingHeaderIsRejected(t *testing.T) {
	rec, seen := serve(t, fakeAuthenticator{}, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if rec.Header().Get("WWW-Authenticate") != authRealm {
		t.Fatalf("expected a WWW-Authenticate challenge, got %q", rec.Header().Get("WWW-Authenticate"))
	}
	if seen != nil {
		t.Fatal("downstream handler should not run without credentials")
	}
}

func TestBasicAuth_InvalidCredentialsLeakNothing(t *testing.T) {
	noHeader, _ := serve(t, fakeAuthenticator{err: ErrInvalidCredentials}, "")
	wrongCreds, _ := serve(t, fakeAuthenticator{err: ErrInvalidCredentials}, basicHeader("a@example.com", "nope"))

	if wrongCreds.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", wrongCreds.Code)
	}
	if wrongCreds.Body.String() != noHeader.Body.String() {
		t.Fatalf("wrong-credentials body %q differs from missing-header body %q (enumeration leak)",
			wrongCreds.Body.String(), noHeader.Body.String())
	}
}

func TestBasicAuth_InfrastructureErrorIsNotA401(t *testing.T) {
	rec, _ := serve(t, fakeAuthenticator{err: errors.New("db down")}, basicHeader("a@example.com", "pw"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for an infrastructure failure, got %d", rec.Code)
	}
}

func TestBasicAuth_SuccessInjectsIdentity(t *testing.T) {
	want := Identity{Subject: uuid.New(), Role: RoleCustomer}
	rec, seen := serve(t, fakeAuthenticator{identity: want}, basicHeader("a@example.com", "pw"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if seen == nil || *seen != want {
		t.Fatalf("downstream identity = %+v, want %+v", seen, want)
	}
}

func TestRequireRoleGuards(t *testing.T) {
	merchant := WithIdentity(context.Background(), Identity{Subject: uuid.New(), Role: RoleMerchant})
	customer := WithIdentity(context.Background(), Identity{Subject: uuid.New(), Role: RoleCustomer})

	if _, err := RequireMerchant(merchant); err != nil {
		t.Fatalf("merchant should pass RequireMerchant: %v", err)
	}
	if _, err := RequireCustomer(merchant); !errors.Is(err, ErrForbidden) {
		t.Fatalf("merchant calling a customer guard should be forbidden, got %v", err)
	}
	if _, err := RequireMerchant(customer); !errors.Is(err, ErrForbidden) {
		t.Fatalf("customer calling a merchant guard should be forbidden, got %v", err)
	}
	if _, err := RequireMerchant(context.Background()); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("no identity should be unauthenticated, got %v", err)
	}
}
