# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Project package layout, developer toolchain (`make tools`), and quality gate (`make check`).
- Fail-fast typed configuration loader and structured (`slog`) logging.
- HTTP server skeleton with a `/healthz` endpoint and graceful shutdown on SIGINT/SIGTERM.
- Pinned `.golangci.yml` lint configuration; CI installs and runs the pinned golangci-lint.
- Database schema as code and Atlas versioned-migration workflow.
- Local `docker-compose` stack (Postgres, Redis, MinIO) wired to `make dev`.
- Pure domain model (entities, enums, typed errors) and repository interfaces.
- Platform helpers: AES-256-GCM encryption, bcrypt hashing, HMAC blind index, keyset pagination cursor, and UUID generation.

- PostgreSQL persistence layer: sqlc queries and repository implementations of the domain
  interfaces, with adapter-side PII encryption, transactional race-safe purchase/checkout,
  keyset pagination, and pgx→domain error mapping. Integration-tested with testcontainers.
- CI codegen-drift gate (regenerates sqlc and fails on any diff).

### Changed
- Adopted the latest stable Go (1.25) and upgraded `golang.org/x/crypto` to v0.53.0 (ADR-0006).
