# 0006 — Adopt the latest stable Go toolchain

- Status: Accepted
- Date: 2026-06-16

## Context

ADR-0001 chose Go 1.24, but the Go version is **not** a challenge constraint — the brief only says "use Golang". Holding 1.24 capped `golang.org/x/crypto` at v0.45.0, because newer releases (v0.50+) require Go 1.25. Pinning a dependency below its latest stable release to preserve a self-imposed Go version is avoidable churn with no functional benefit.

## Decision

Track the **latest stable Go** (currently 1.25.x) and the latest stable dependencies. `go.mod` declares `go 1.25.0`; CI and the release workflow build on Go 1.25; `golang.org/x/crypto` is upgraded to v0.53.0. This supersedes the Go-version choice in [ADR-0001](0001-backend-stack.md) (the rest of that stack decision stands).

## Consequences

- No artificial dependency pinning — the project stays on current security and maintenance lines.
- The `go.mod` `go` directive and CI's `go-version` move forward together with each Go release.
- A future Go bump is a one-line `go.mod` + CI change, not a new ADR, unless the *policy* itself changes.
