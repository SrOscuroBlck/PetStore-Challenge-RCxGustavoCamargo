---
name: manage-db-schema
description: Use when changing the database schema — a new table, column, index, enum value, or constraint. Walks the Atlas versioned-migration flow (edit schema, diff, lint, apply, regenerate) and keeps DATA_MODEL.md in sync, so schema changes are safe, reviewable, and documented.
---

# Manage the database schema

The schema is code, managed by Atlas. Migrations are versioned, linted, and immutable once created. sqlc reads the schema as its source of truth, so a schema change is always followed by regeneration.

Prerequisite: `make tools` (installs Atlas).

## 1. Edit the schema source

Change the Atlas schema under `db/schema/`. Apply the data-model rules: fixed value-sets are enum types; don't add a column for a value you can derive (except historical facts frozen in time); index the access paths used by the hot read queries and keyset pagination (`(store_id, created_at, id)` style); sensitive columns are stored encrypted (see SECURITY.md).

## 2. Generate a versioned migration

```bash
make migrate-new name=<short_snake_case>   # atlas migrate diff
```

Atlas diffs the schema and writes a new migration file under `db/migrations/`. **Never hand-edit an existing migration** — they are immutable once created; correct a mistake with a new migration. (A hook blocks edits to `db/migrations/`.)

## 3. Lint the migration

```bash
make migrate-lint    # atlas migrate lint
```

Resolve anything flagged as destructive or unsafe (dropping a column with data, a non-concurrent index on a large table, etc.) before applying.

## 4. Apply

```bash
make migrate-up
```

## 5. Regenerate dependent code

The schema changed, so regenerate type-safe queries:

```bash
make generate    # sqlc (and gqlgen if the API shape changed)
```

## 6. Update the standard doc

`docs/DATA_MODEL.md` is a standard the rest of the code builds toward — keep its ERD, index list, and encryption table in sync with the migration. If the implementation **deviated** from what the doc described, update the doc **and state why** (or write a new ADR if it's a real decision change). Never let the doc and the schema disagree.

## 7. Verify

`make check` passes; the migration applies cleanly on a fresh database (the integration-test setup runs migrations from scratch — see **add-integration-test**).
