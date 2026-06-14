---
name: add-integration-test
description: Use when adding tests for a repository or a concurrency-sensitive use case (purchase, checkout, removal). Encodes the testcontainers setup against a real Postgres/Redis and the race-proof pattern that demonstrates exactly one winner under concurrent access — the core correctness requirement of this challenge.
---

# Add an integration test

Exhaustive unit coverage is explicitly optional for this challenge. Spend the testing effort where it proves the requirements: **repository correctness against a real database** and **race-safety under concurrency**.

Prerequisite: Docker running (testcontainers-go starts real Postgres/Redis).

## Real dependencies, not mocks

- Use `testcontainers-go` to start Postgres (and Redis where the path uses cache).
- In `TestMain`, start one container per package, run the Atlas migrations against it from scratch, and share the connection. Tear down at the end.
- Testing against the real engine is the point — it catches constraint, transaction, and SQL-dialect issues a mock never would.

## The race-proof pattern (the heart of the challenge)

For every purchase path, prove that concurrency cannot double-sell:

1. Seed one `AVAILABLE` pet.
2. Launch N goroutines, each calling the purchase use case, released together with a start barrier (`sync.WaitGroup` / a closed channel) so they truly contend.
3. Collect results and assert: **exactly one** succeeds; the other N-1 return the typed `ErrPetAlreadySold`; the pet ends `SOLD` exactly once with a single buyer recorded.

Apply the same shape to:
- **Cart checkout overlap** — two carts sharing a pet: exactly one checkout succeeds, the other rolls back fully and names the unavailable pet.
- **Remove vs purchase** — a removal and a purchase racing the same pet: exactly one wins.

## Conventions

- Always run with the race detector: `go test -race ./...` (and via `make check`).
- Name tests for the behavior, not the function: `TestCheckout_RollsBackWhenAnyPetSold`, `TestPurchasePet_OnlyOneWinnerUnderConcurrency`.
- Table-driven where cases share a shape; keep each assertion about one behavior.
- Tests read fresh state from the database to verify the final outcome — don't assert on an object captured before the operation.

## Verify

The new test passes under `-race`, and fails if the conditional-write / locking guard is removed (a race test that still passes without the guard isn't testing the race).
