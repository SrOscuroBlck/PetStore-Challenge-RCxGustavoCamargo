# 0001 — Backend stack

- Status: Accepted (Go-version choice superseded by [0006](0006-go-version-policy.md))
- Date: 2026-06-13

## Context

The challenge allows Go or Rust for the backend, requires PostgreSQL, Redis, and a local Docker + Minikube deployment, and forbids externally-hosted services. This repository is scoped to the backend only.

## Decision

Use **Go 1.24** as the implementation language, with this stack:

- **GraphQL** (`gqlgen`) as the API protocol — see [0002](0002-graphql-over-grpc.md).
- **PostgreSQL** via `pgx/v5`, with **sqlc** and **Atlas** — see [0003](0003-sqlc-and-atlas.md).
- **Redis** (`go-redis`) as a read cache.
- **MinIO** for pet pictures — see [0005](0005-image-storage-minio.md).
- **Docker + Minikube** for the local cluster.
- **bcrypt** for password hashing; app-layer **AES-256-GCM** for PII at rest.

## Consequences

- Go's first-class concurrency and the maturity of `pgx`/`gqlgen` fit a read-heavy, concurrency-sensitive API well.
- The whole stack is self-hostable in Minikube with no external dependencies, satisfying the "runnable locally, no hosted services" constraint.
- A clean layered architecture keeps these infrastructure choices swappable behind interfaces.
