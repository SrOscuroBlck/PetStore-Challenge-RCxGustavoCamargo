---
name: document-on-build
description: Keep documentation in lockstep with the code. Use AFTER implementing or changing something that the docs describe — a GraphQL query/mutation/subscription, a run/config/test capability, or a design decision. Documents what now exists; never predicts what doesn't.
---

# Document on build

The first failure mode of documentation is writing it before the code. Anticipatory docs are wrong the moment the implementation makes a different choice, and they erode trust in every other doc. This skill enforces the rule: **document a thing only once it exists, and update its doc in the same change that alters it.**

## Two kinds of documentation

**Standard docs** — the design you build *toward*. Set up front; they constrain implementation.
- `docs/ARCHITECTURE.md`, `docs/DATA_MODEL.md`, `docs/SECURITY.md`, `docs/adr/*`
- Rule: if your implementation **deviates** from one of these, you do not silently diverge. You update the doc to match reality **and record why** (a note in the doc, or a new superseding ADR for a real decision change). The doc stays the source of truth a future dev can rely on.

**Reference docs** — the surface you *have already built*. Written after the fact.
- `docs/API.md`, and the operational sections of `README.md` (running, configuration, testing, usage)
- Rule: never write an entry for something that isn't implemented yet. Add the entry in the same change that adds the code; remove it when the code is removed.

## When this skill triggers

Invoke it as part of finishing a unit of work, before you call the task done:

- Added or changed a **GraphQL operation** → update `docs/API.md` (see below).
- Made the system **runnable / configurable / testable** in a new way (a real `make` target, a required env var, a working local/Minikube path) → update the matching `README.md` section. Only document commands that actually work now.
- **Deviated from a standard doc** → update that doc and explain the deviation.

## Updating `docs/API.md`

For each query / mutation / subscription you implemented or changed, record exactly what the code now does:

- Operation name and kind (query / mutation / subscription)
- Required role / auth (merchant, customer)
- Arguments (with types) and pagination shape if it's a list
- Return type / shape
- Error codes it can raise, with what triggers each
- One concrete, copy-pasteable example for the operation

Keep examples consistent with the actual schema. If you renamed or removed an operation, fix or delete its entry in the same change — a stale example is worse than no example.

## Checklist before marking work done

- [ ] Did I add/change a GraphQL operation? → `docs/API.md` reflects it, with an example.
- [ ] Did I add a real way to run/configure/test the system? → `README.md` documents it, and the commands actually work.
- [ ] Did I diverge from `ARCHITECTURE.md` / `DATA_MODEL.md` / `SECURITY.md` / an ADR? → updated it and stated why.
- [ ] Did I document anything that does **not** yet exist in the code? → remove it.
