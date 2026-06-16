# 0004 — Database-enforced race-condition strategy

- Status: Accepted
- Date: 2026-06-13

## Context

Two customers must never buy the same pet, and a merchant removal can race a purchase. Coordinating this in application code is fragile — it breaks the moment more than one API instance runs.

## Decision

Enforce race-safety at the database, never with application-level coordination:

- **Single purchase / removal** — one atomic conditional write, `UPDATE pets SET ... WHERE id = $1 AND status = 'AVAILABLE'`; the winner is decided by rows-affected. No read-then-write window.
- **Cart checkout** — a single transaction that locks the target rows `SELECT ... FOR UPDATE` in deterministic id order (deadlock-free), verifies availability inside the transaction, and is all-or-nothing; the error names every unavailable pet.
- **Idempotency** — a pet already sold to the *same* customer is a success, not a second sale.
- Availability is read **inside the transaction**, never from a pet fetched earlier in the request.

## Consequences

- Correctness rests on PostgreSQL guarantees and holds under concurrency and horizontal scaling.
- `pets.status` is a stored test-and-set column rather than a derived value (justified in `DATA_MODEL.md`).
- The in-memory domain transitions (`Pet.MarkSold` / `MarkRemoved`) are pure state changes, **not** the concurrency gate — the conditional write is.
