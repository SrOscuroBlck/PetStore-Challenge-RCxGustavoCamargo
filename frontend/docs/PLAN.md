# Build plan

The customer storefront, planned as shippable phases. Each phase is independently demonstrable. The
acceptance criteria below trace directly to the **customer user stories** in `docs/challenge.md` — the
demo and grading map onto these.

## User stories → acceptance criteria

The challenge defines exactly four customer stories. These are the definition of done:

1. **Browse a store's available pets**
   - Opening `/store/:storeId` shows that store's `AVAILABLE` pets, with name, species, age, description,
     and image.
   - Purchased/removed pets never appear; a browser refresh reflects current availability.
   - Unknown store → friendly empty state; loading → skeletons; error → banner with retry.

2. **Instant single purchase**
   - A Purchase button on each pet buys it instantly and removes it from the catalog (optimistic).
   - If the pet is no longer available, a **human-readable** error is shown and the UI reconciles to
     server state (the pet is gone / marked sold).

3. **Add to cart**
   - An Add-to-cart button adds a pet to a cart with a visible count and a reviewable list; items can be
     removed.

4. **Cart checkout**
   - A Checkout button purchases all cart items atomically.
   - On failure, a **human-readable** error names **every pet no longer available** (server message
     surfaced verbatim); the cart reconciles.

Cross-cutting (from Requirements): pagination on the catalog; <2s under load via pagination + cache +
lazy images; security per `docs/THREAT_MODEL.md`; race conditions handled per ADR-0004; polished,
accessible, responsive UI; one documented local run path.

## Phases

| # | Phase | Outcome | Status |
|---|---|---|---|
| 0 | Framework & docs | `.claude/` guardrails, CLAUDE.md, ADRs, threat model, this plan | ✅ Done |
| 1 | Scaffold & data layer | Vite+TS strict, Tailwind, Router, Apollo (auth + error link + relay pagination), graphql-codegen wired to `docs/schema.graphqls`; health-check against running backend | ✅ Done |
| 2 | Store entry | Open storefront (no login wall); auth injected at the gateway (ADR-0005); `/store/:storeId` routing + app shell | ✅ Done |
| 3 | Catalog (Story 1) | `availablePets` infinite scroll, pet cards, skeleton/empty/error states, image fallbacks | ✅ Done |
| 4 | Single purchase (Story 2) | Optimistic `purchasePet` + rollback + `UNAVAILABLE` toast | ✅ Done |
| 5 | Cart & checkout (Stories 3–4) | Cart state + count + reviewable list; atomic `checkout`; success animation + named-unavailable-pets error | ✅ Done |
| 6 | Polish & a11y | Page/list transitions, `prefers-reduced-motion`, dark mode, focus management, error boundary, CSP headers | ✅ Done |
| 7 | Infra & run path | Dockerfile, K8s manifests, same-origin ingress routing, one-command up, grader-blind README (ADR-0003) | ✅ Done |
| 8 | Tests | Critical flows: auth, browse/pagination, purchase rollback, checkout named-errors, error mapping | ✅ Done |

## Tooling support per phase

- **Phase 1, 3–5:** use the `new-operation` skill for every GraphQL op; run `contract-checker` on new
  documents.
- **Phase 3–6:** use the `new-component` skill; run `ux-state-auditor` on each screen.
- **Every phase:** the `guard.js` / `verify-changes.js` / `quality-gate.js` hooks run automatically;
  use the `quality-gate` skill before declaring a phase done.

## Definition of done (per phase)

Typecheck + lint clean; the phase's acceptance criteria demonstrable in the running app; new operations
contract-checked; new screens state-audited; no guardrail escape hatches added without a noted reason.
