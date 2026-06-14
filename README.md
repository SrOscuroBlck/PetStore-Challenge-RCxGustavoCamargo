# Pet Store — Backend

A backend for a multi-tenant pet store. **Merchants** manage their store's pet listings through a GraphQL API; **customers** browse available pets and purchase them individually or via a cart checkout. Built for the Robotic Crew challenge (`docs/challenge.md`).

This repository is the **backend only** and runs entirely on your machine — no externally-hosted services.

> **Status:** early development. This README documents what has been **decided** (stack and architecture). Operational sections — how to run it, configure it, and call the API — are added to this file and to [`docs/API.md`](docs/API.md) **as each piece is actually built**, so the documentation always reflects the code rather than predicting it.

---

## Tech stack

| Concern | Choice |
|---|---|
| Language | Go 1.24 |
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

## Documentation

| Document | What's inside |
|---|---|
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | Layering, request lifecycle, concurrency & race-condition strategy |
| [`docs/DATA_MODEL.md`](docs/DATA_MODEL.md) | Entity-relationship diagram, tables, indexes, what's encrypted |
| [`docs/SECURITY.md`](docs/SECURITY.md) | Authentication, store isolation, encryption, hardening, TLS |
| [`docs/API.md`](docs/API.md) | GraphQL operations reference — populated as each operation is implemented |
| [`docs/adr/`](docs/adr/) | Architecture Decision Records (the "why" behind key choices) |
| [`docs/challenge.md`](docs/challenge.md) | The original challenge brief |
