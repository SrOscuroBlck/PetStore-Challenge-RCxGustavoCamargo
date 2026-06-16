package auth

import "errors"

// ErrUnauthenticated means no identity was established for the request.
// ErrInvalidCredentials is returned for every credential failure mode without
// revealing which factor failed, so it cannot be used to enumerate accounts.
// ErrForbidden means the authenticated role may not invoke this operation.
var (
	ErrUnauthenticated    = errors.New("auth: not authenticated")
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	ErrForbidden          = errors.New("auth: forbidden for this role")
)
