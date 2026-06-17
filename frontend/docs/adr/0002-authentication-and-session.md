# ADR-0002: Authentication & session

**Status:** Superseded by [ADR-0005](0005-open-storefront-gateway-auth.md) · **Date:** 2026-06-17

> **Superseded.** This ADR specified a login screen with a client-held credential. We reconsidered:
> the customer stories ask for a products-first storefront and the domain has no payment, so we moved
> authentication to the same-origin gateway and removed the login wall. See ADR-0005 for the reasoning.
> The text below is kept for decision history.

## Context

The backend uses **HTTP Basic auth on every request** — no login/signup mutation, no token, no session
endpoint (`docs/BACKEND_INTEGRATION.md` §2). The frontend must present a login experience anyway, attach
credentials to every `/graphql` call, and decide where the credential lives. Storage choice is a direct
security trade-off the challenge calls out (sensitive data must not be exposed; common web vectors must
be addressed).

## Decision

- **Credential = `base64(email:password)`**, attached as `Authorization: Basic …` to every GraphQL
  request via an Apollo auth link.
- **"Login" is login-by-probe.** There is no auth endpoint, so we collect email + password and validate
  by firing one cheap authed query (`availablePets` with `first: 1`). HTTP 401 / `UNAUTHENTICATED` =
  bad credentials; success = authenticated. No credential is accepted into the app until it validates.
- **Storage: `sessionStorage`, never `localStorage`.** The Basic value is held in a React auth context,
  mirrored to `sessionStorage` so an in-tab refresh keeps the user logged in (re-validated by one probe
  on load). It is cleared on logout and naturally dies when the tab closes.
- **Customer-only.** Merchant operations return `FORBIDDEN` and are never called or surfaced.

## Consequences

- A full-page refresh stays logged in within the tab (smooth demo); closing the tab logs out.
- The credential is exposed to JS, so XSS is the principal threat — addressed in `docs/THREAT_MODEL.md`
  (no `dangerouslySetInnerHTML` on untrusted data — guardrail-enforced —, CSP, dependency hygiene).
- `UNAUTHENTICATED` anywhere in the app routes back to login and clears the stored credential.

## Alternatives considered

- **In-memory only (no persistence).** Strictest posture, but any refresh forces re-login — poor for a
  graded demo where refresh is the documented way to see sold pets disappear. Rejected on UX; the XSS
  delta vs. sessionStorage is small given our XSS mitigations.
- **`localStorage`.** Survives across tabs/restarts but is the canonical XSS-exfiltration target for
  credentials. Explicitly forbidden by the brief and blocked by a guardrail.
- **Backend-issued session cookie / JWT.** Would be cleaner, but there is no such endpoint and the
  contract is fixed — out of scope.
