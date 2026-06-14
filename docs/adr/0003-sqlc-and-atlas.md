# 0003 — sqlc for data access, Atlas for migrations

- Status: Accepted
- Date: 2026-06-13

## Context

The backend needs type-safe database access that keeps SQL explicit and out of the business layer, plus a serious, industry-standard schema/migration workflow — not a minimal one. The data layer must remain behind repository interfaces.

## Decision

- **sqlc** generates type-safe Go from hand-written SQL in `db/queries/`. Generated code is wrapped by repository implementations in `internal/adapter/postgres`.
- **Atlas** manages the schema as code and produces versioned, linted migrations in `db/migrations/`. sqlc reads the Atlas-managed schema as its source of truth.

## Consequences

- SQL stays explicit and reviewable; no reflection-based ORM hides query behavior or couples business logic to the database (rules 7 and 13).
- Atlas adds schema diffing, migration linting (catching unsafe changes), and drift detection — stronger guarantees than a plain up/down-file migrator. It runs as an init/job step in Minikube.
- Two code-generation steps (`sqlc generate`, plus `gqlgen generate` for the API) are wired into `make generate`; contributors must regenerate after schema changes.
- The pairing requires Atlas and sqlc to agree on the schema; the Atlas schema is the canonical input to both.
