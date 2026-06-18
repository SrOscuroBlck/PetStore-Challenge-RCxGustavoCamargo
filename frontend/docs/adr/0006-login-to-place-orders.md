# ADR-0006: Login to place orders (open browse, gated ordering)

**Status:** Accepted · **Date:** 2026-06-17 · **Amends:** [ADR-0005](0005-open-storefront-gateway-auth.md)

## Context

The challenge overview says *"Customers will be able to **login** to the store and **place orders**."*
ADR-0005 chose an open storefront with no login at all (gateway injects an ambient credential). That
keeps browsing friction-free, but it drops the "login" the brief calls for. We reconcile the two: **keep
browsing open, require a login to place an order** — the realistic storefront pattern (you browse freely;
you sign in to check out) and the literal reading of "login … place orders".

`availablePets` requires auth, so browsing cannot be truly anonymous against this backend; an *ambient*
credential injected at the gateway covers reads.

## Decision

- **Browse is open** — no login; the gateway injects an ambient browse credential (ADR-0005) for
  `availablePets`/images.
- **Placing an order requires sign-in.** Buy and Checkout call `ensureSignedIn()`, which opens a login
  dialog when signed out and resumes the action on success. Credentials are validated by **login-by-probe**
  (a cheap `availablePets(first:1)` with the candidate `Authorization`) before being stored.
- **The signed-in credential is attached client-side** (Apollo `authLink`) to subsequent requests. The
  **gateway passes a client `Authorization` through, injecting the ambient credential only when none is
  present** — so orders are attributed to the signed-in customer. Dev: the Vite proxy does this
  conditional injection; prod: the nginx gateway will (deployment).
- **Credential storage:** in memory + `sessionStorage` (survives in-tab refresh, dies on tab close) —
  **never `localStorage`**; cleared on logout.

## Consequences

- Satisfies the brief's "login and place orders" while keeping a products-first, open catalog.
- The customer credential now lives in the browser (`sessionStorage`) **for the duration of an order
  session** — a change from ADR-0005's "never in the browser". XSS mitigations (no
  `dangerouslySetInnerHTML`, CSP at serve-time, dependency hygiene) carry the risk; see
  `docs/THREAT_MODEL.md`.
- `UNAUTHENTICATED` on an order routes back to the login dialog.
- The gateway gains one rule (conditional injection / pass-through); otherwise the same-origin topology
  (ADR-0003) is unchanged.

## Alternatives considered

- **No login at all (ADR-0005).** Cleanest security posture (no client credential) but omits the brief's
  login; amended here for the order path.
- **Full login wall (ADR-0002).** Gates browsing too — rejected; the catalog should be open.
