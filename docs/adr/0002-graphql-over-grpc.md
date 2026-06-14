# 0002 — GraphQL as the protocol; no gRPC

- Status: Accepted
- Date: 2026-06-13

## Context

The challenge requires GraphQL or gRPC Web (REST and plain HTTP are disallowed). The domain is a small set of operations: merchant pet CRUD plus two filtered list queries, and customer browse/purchase/checkout. Reads dominate and need pagination and caching.

## Decision

Use **GraphQL** (schema-first via `gqlgen`) as the sole API protocol. Do **not** introduce gRPC.

## Consequences

- A single typed schema serves both merchant and customer surfaces, with enums, cursor-connection pagination, and a self-documenting contract out of the box.
- gRPC's strengths (streaming, strict service-to-service contracts, low-latency internal RPC) are not needed for a single client-facing API over a handful of operations; adding it would be scope without justification.
- Should a concrete need for gRPC appear later (e.g. an internal high-throughput service boundary), it would warrant a new ADR superseding this one.
