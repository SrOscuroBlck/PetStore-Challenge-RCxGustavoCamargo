# ADR-0001: Frontend stack & GraphQL client

**Status:** Accepted · **Date:** 2026-06-17

## Context

The challenge mandates React + TypeScript and a typed GraphQL protocol; it leaves build tooling,
routing, styling, and the specific GraphQL client open ("open to our choices, potential bonus points").
The backend contract is fixed and ships a GraphQL schema (`docs/schema.graphqls`), so the client must
support code generation of types and typed hooks. The catalog is a Relay cursor connection and purchases
are race-prone, so the client's pagination and optimistic-update story matter materially.

## Decision

- **React + TypeScript (strict)** — mandated; strict mode, no `any` on API data.
- **Vite** for dev/build and **React Router** for routing — fast HMR, first-class code-splitting,
  standard SPA routing for `/store/:storeId`.
- **Apollo Client** as the typed GraphQL client, with **graphql-codegen** generating types and typed
  hooks from `docs/schema.graphqls`. No hand-written, untyped operations.
- **Tailwind CSS** + **Framer Motion** for styling and animation.

### Why Apollo over urql

Both satisfy "typed client + codegen." Apollo wins here on the two axes this app stresses:

1. **Relay cursor pagination** — Apollo's `relayStylePagination` cache helper handles `edges`/`pageInfo`
   merging out of the box, which is the catalog's core read pattern (infinite scroll).
2. **Optimistic mutations + rollback** — purchase/checkout need optimistic UI that *rolls back* on a
   server `UNAVAILABLE`. Apollo's `optimisticResponse` + cache `update` + automatic error rollback maps
   directly onto this; urql requires more manual wiring via graphcache exchanges.

The cost is a larger bundle than urql. We accept it: it's mitigated by route code-splitting, and the
normalized cache reduces refetching (a net win for the <2s / 1k-concurrent target).

## Consequences

- The Apollo normalized cache is the source of truth for server data; local state is limited to cart and
  session credential.
- All operations live in `.graphql` documents and flow through generated hooks; a hand-written query is
  a process violation (enforced by the `new-operation` skill and `contract-checker` agent).
- Bundle size must be watched; routes are code-split and images lazy-loaded.

## Alternatives considered

- **urql** — lighter, but weaker batteries-included pagination/optimistic story for our exact patterns.
- **Relay** — most rigorous for cursor connections, but heavier conceptual overhead and stricter schema
  conventions than this small, fixed contract warrants.
- **TanStack Query + graphql-request** — great for REST-ish flows, but gives up the normalized GraphQL
  cache that makes catalog freshness and optimistic updates clean.
