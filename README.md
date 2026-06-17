# Pet Store — Backend

A backend for a multi-tenant pet store. **Merchants** manage their store's pet listings through a GraphQL API; **customers** browse available pets and purchase them individually or via a cart checkout. Built for the Robotic Crew challenge (`docs/challenge.md`).

This repository is the **backend only** and runs entirely on your machine — no externally-hosted services.

> **Status:** early development. This README documents what has been **decided** (stack and architecture). Operational sections — how to run it, configure it, and call the API — are added to this file and to [`docs/API.md`](docs/API.md) **as each piece is actually built**, so the documentation always reflects the code rather than predicting it.

---

## Tech stack

| Concern | Choice |
|---|---|
| Language | Go 1.25 |
| API | GraphQL (`gqlgen`, schema-first) |
| Database | PostgreSQL (`pgx/v5`) with type-safe queries via `sqlc` |
| Schema & migrations | Atlas (versioned, linted) |
| Cache | Redis (`go-redis`) |
| Object storage | MinIO (self-hosted, S3-compatible) for pet pictures |
| Auth | HTTP Basic + bcrypt password hashing |
| Orchestration | Docker + Minikube (local Kubernetes) |

The reasoning behind each choice is recorded in [`docs/adr/`](docs/adr/).

---

## Architecture

The API is layered so business rules never depend on infrastructure: a pure domain core, application services that own use cases and transaction boundaries, and adapters that implement persistence, caching, and object storage behind interfaces. Multi-tenant isolation, race-safe purchasing, cached reads, and encryption in transit and at rest are designed in from the start.

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for the full picture.

---

## Development

Requires Go 1.25+ and `make`.

```bash
make tools                 # install gqlgen, sqlc, atlas, goimports, golangci-lint into ./bin
make check                 # format check, go vet, golangci-lint, build, race tests
```

Local dependencies and database (requires Docker):

```bash
cp .env.example .env   # then set PII_ENCRYPTION_KEY (see below); the Makefile auto-loads .env
make dev               # start Postgres, Redis, MinIO via docker-compose
make migrate-up        # apply the database schema
make tls-certs         # generate a local self-signed TLS cert into ./certs
make run               # run the server over TLS; GET /healthz returns {"status":"ok"}
```

`make run` serves **HTTPS only** (plaintext is refused) and connects to Postgres, MinIO, and Redis,
reading `PII_ENCRYPTION_KEY`, `REDIS_ADDR`, the `MINIO_*` vars, and `TLS_CERT_FILE`/`TLS_KEY_FILE`,
so it needs `make dev`, `make tls-certs`, and a populated `.env`. Generate the encryption key once
with `openssl rand -base64 32`; the MinIO bucket is created automatically at startup, and the
catalog cache falls back to Postgres if Redis is unavailable. Hit it with a self-signed cert via
`curl -k https://localhost:8443/healthz`. The GraphQL endpoint at `/graphql` is behind HTTP Basic
auth (requests without valid credentials get 401); schema introspection is off unless
`GRAPHQL_INTROSPECTION=true`.

Configuration is read from environment variables and validated at startup — a missing
required value aborts with an error naming it. See [`.env.example`](.env.example) for the
variables currently in use (more are added as features that need them land).

---

## Documentation

| Document | What's inside |
|---|---|
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | Layering, request lifecycle, concurrency & race-condition strategy |
| [`docs/DATA_MODEL.md`](docs/DATA_MODEL.md) | Entity-relationship diagram, tables, indexes, what's encrypted |
| [`docs/SECURITY.md`](docs/SECURITY.md) | Authentication, store isolation, encryption, hardening, TLS |
| [`docs/API.md`](docs/API.md) | GraphQL operations reference — populated as each operation is implemented |
| [`docs/adr/`](docs/adr/) | Architecture Decision Records (the "why" behind key choices) |
| [`docs/challenge.md`](docs/challenge.md) | The original challenge brief |
