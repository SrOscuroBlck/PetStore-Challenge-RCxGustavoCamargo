---
name: add-persistence-query
description: Use when an app service needs a new database read or write. Walks the sqlc-backed flow — write SQL, generate, declare the repository interface in the consuming layer, implement the adapter, map errors to domain types — so persistence stays behind interfaces and never leaks into business logic.
---

# Add a persistence query

Data access lives behind repository interfaces; business logic never sees SQL or a connection. This skill keeps that boundary intact.

Prerequisite: `make tools` (installs sqlc). Generated files are off-limits to hand edits (a hook blocks it).

## 1. Check what exists first

Spawn **dependency-lookup**: a repository method that does this fetch/write may already exist. Don't add a second.

## 2. Make sure the schema supports it

If the query needs a column, table, or index that doesn't exist yet, use the **manage-db-schema** skill before writing the query.

## 3. Write the SQL

Add the query to `db/queries/*.sql` with a sqlc annotation and **parameters only** — never string-built SQL:

```sql
-- name: MarkPetSold :execrows
UPDATE pets SET status = 'SOLD', sold_at = now(), sold_by_customer_id = $2
WHERE id = $1 AND status = 'AVAILABLE';
```

Use the right return mode (`:one`, `:many`, `:exec`, `:execrows`). For race-safe writes prefer conditional `UPDATE ... WHERE status = 'AVAILABLE'` and inspect rows-affected; for multi-row atomic operations lock with `SELECT ... FOR UPDATE` ordered by id inside a transaction.

## 4. Generate

`make generate` (sqlc). Leave the generated code untouched.

## 5. Declare the repository interface

In the layer that **uses** it (`internal/domain` or `internal/app`), add the method to the repository interface. It speaks **domain types**, returns `(T, error)`, and never exposes sqlc/pgx types to callers.

## 6. Implement the adapter

In `internal/adapter/postgres`, implement the method by wrapping the generated query:

- Map sqlc rows ↔ domain types explicitly.
- Translate driver errors into **typed domain errors**: `pgx.ErrNoRows` → domain not-found; unique-violation (`23505`) → domain conflict. Callers use `errors.Is`/`errors.As`, never inspect pgx codes.
- Use the injected `*pgxpool.Pool` (built once at startup). For atomic multi-statement work, run inside a transaction.

## 7. Verify

`make check` passes (incl. `go test -race`). Add or extend an integration test with the **add-integration-test** skill — repository correctness is tested against a real Postgres, not a mock. Run **clean-code-reviewer**, and **security-concurrency-auditor** if the query is on a purchase/removal path.
