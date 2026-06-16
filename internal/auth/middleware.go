package auth

import (
	"context"
	"errors"
	"net/http"
)

const authRealm = `Basic realm="petstore"`

type authenticate interface {
	Authenticate(ctx context.Context, email, password string) (Identity, error)
}

// BasicAuth parses HTTP Basic credentials, resolves them to an Identity, and
// stores that identity in the request context for downstream handlers. Requests
// without valid credentials are rejected with 401 before any handler runs. The
// response never names the failed factor, so it cannot be used to enumerate
// accounts, and credentials are never logged.
func BasicAuth(authenticator authenticate) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			email, password, ok := r.BasicAuth()
			if !ok {
				unauthorized(w)
				return
			}
			identity, err := authenticator.Authenticate(r.Context(), email, password)
			if err != nil {
				if errors.Is(err, ErrInvalidCredentials) {
					unauthorized(w)
					return
				}
				http.Error(w, "authentication failed", http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithIdentity(r.Context(), identity)))
		})
	}
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", authRealm)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}

// RequireMerchant returns the identity only when the request was authenticated
// as a merchant; it is the guard a merchant operation calls before doing work.
func RequireMerchant(ctx context.Context) (Identity, error) {
	return requireRole(ctx, RoleMerchant)
}

// RequireCustomer is the customer-side counterpart to RequireMerchant.
func RequireCustomer(ctx context.Context) (Identity, error) {
	return requireRole(ctx, RoleCustomer)
}

func requireRole(ctx context.Context, role Role) (Identity, error) {
	identity, ok := IdentityFromContext(ctx)
	if !ok {
		return Identity{}, ErrUnauthenticated
	}
	if identity.Role != role {
		return Identity{}, ErrForbidden
	}
	return identity, nil
}
