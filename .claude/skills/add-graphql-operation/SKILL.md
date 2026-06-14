---
name: add-graphql-operation
description: Use when adding or changing a GraphQL query, mutation, or subscription. Walks the full vertical slice in order — schema, codegen, app service, thin resolver, auth/isolation, error mapping, docs, review — so the operation lands consistently across every layer instead of being half-wired.
---

# Add a GraphQL operation

A new operation touches several layers in a fixed order. Skipping a step is how resolvers end up with business logic, or schemas drift from resolvers. Follow the slice top to bottom.

Prerequisite: code generation tools are installed (`make tools` installs gqlgen/sqlc/atlas/goimports/golangci-lint). Never hand-edit generated files — a hook blocks it.

## 1. Check what exists first

Spawn the **dependency-lookup** agent before writing anything: the type, enum, error, or app method you need may already exist. Reuse beats recreate.

## 2. Define it in the schema

Edit the schema under `internal/graph/schema/*.graphqls`. Conventions:

- Mutations take a single `Input` type and **return the affected entity** (so the client can read back the new state).
- List queries are **Relay cursor connections**: `first` / `after` arguments, `edges { node cursor }` / `pageInfo { hasNextPage endCursor }`. No offset arguments.
- Fixed value-sets are `enum`s, never free strings.
- Nullability is deliberate — non-null by default; nullable only when "absent" is a real, meaningful state.
- Fields visible only to one role (e.g. breeder details) are documented as such; enforcement happens in the resolver/service, not the schema.

## 3. Generate

Run `make generate` (gqlgen). This regenerates models and resolver stubs. Do not touch the generated output.

## 4. Implement the resolver — thin

The resolver is a table of contents, not the prose:

1. Read the authenticated identity from `context.Context` (never from arguments).
2. Parse/validate the input into a typed command; reject bad input fast.
3. Call the **app service** for the use case.
4. Map the domain result to the GraphQL type.
5. Translate any domain error into a GraphQL error with a stable `code` extension and a human-readable message — this is the **only** place errors are translated.

No business logic, no SQL, no repository or pool access in the resolver.

## 5. Implement / extend the app service

The use case lives in `internal/app`. It owns the transaction boundary, applies domain rules, fetches fresh state, and decides cache reads/invalidation. If it needs new persistence, use the **add-persistence-query** skill. Keep purchase/availability decisions race-safe (conditional writes / `FOR UPDATE`).

## 6. Enforce security

Confirm role separation (right role for this operation) and store isolation (merchant scope derived from identity, never a client-supplied id).

## 7. Document it

Use the **document-on-build** skill to record the operation in `docs/API.md` — name, kind, role, arguments, return, error codes, and one working example.

## 8. Verify

- `make generate` is clean and committed alongside the source.
- `make check` passes (fmt, vet, lint, build, `go test -race`).
- Run the **clean-code-reviewer** agent; if the operation touches purchasing, auth, or sensitive data, also run **security-concurrency-auditor**.
