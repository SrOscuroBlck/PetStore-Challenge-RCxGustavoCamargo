# ADR-0004: Catalog freshness & optimistic concurrency

**Status:** Accepted · **Date:** 2026-06-17

## Context

Availability is **server-derived** and pets are contended: two customers can race the same pet and
exactly one wins (`docs/API.md`, `docs/BACKEND_INTEGRATION.md` §7). The user stories require:
immediate purchase feedback; a human-readable error when a pet is no longer available; a cart checkout
that, on failure, **names every pet that is no longer available**; and that a browser refresh reflects
sold/removed pets (sold pets must not appear in the catalog).

This is the app's central correctness concern, so the strategy is recorded explicitly.

## Decision

- **Server is the source of truth for availability.** The catalog queries `availablePets`, which returns
  only `AVAILABLE` pets; we never client-side resurrect or guess availability. A refetch/refresh
  therefore drops sold/removed pets automatically.
- **Optimistic purchase, with rollback.** `purchasePet` applies an optimistic update (the pet animates
  out of the catalog immediately). If the server returns `UNAVAILABLE`, Apollo rolls the cache back and
  we surface the human-readable message via a toast, leaving the catalog consistent with the server.
- **Atomic checkout, named failures.** `checkout` is all-or-nothing. On `UNAVAILABLE`, we show the
  **server's message verbatim** because it already names the unavailable pets (a hard challenge
  requirement) — we do not invent our own text. On success, the purchased pets animate out and the cart
  clears.
- **Branch on `extensions.code`, display `message`.** Behavior keys off the stable error code; the
  user sees the backend's human-readable message. `UNAUTHENTICATED` → re-auth; `NOT_FOUND` → "no longer
  listed"; `INTERNAL`/`VALIDATION` → generic friendly fallback.
- **Freshness on focus.** The catalog refetches on window focus / load so a returning user sees current
  availability without a manual refresh (refresh alone is sufficient per the brief; this is polish).

## Consequences

- The UI never assumes a locally-shown pet is still buyable; the authoritative answer is the mutation
  result.
- Optimistic logic lives in the mutation hooks (cache `update` + `optimisticResponse` + rollback), not
  scattered in components — components render cache state.
- No `SOLD`/`REMOVED` pet is ever rendered in the catalog list.

## Alternatives considered

- **Pessimistic UI (await server before any visual change).** Simpler, but loses the "instant"
  feedback the stories ask for and feels sluggish under latency. Rejected.
- **Client-side polling/subscriptions for availability.** The contract offers no subscription; polling
  adds load against the <2s/1k-concurrent target. Refetch-on-focus + the mutation's own result is
  sufficient. Rejected for now.
