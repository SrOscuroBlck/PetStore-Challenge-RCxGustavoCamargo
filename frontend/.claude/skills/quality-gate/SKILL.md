---
name: quality-gate
description: Run the full pre-submission quality gate — typecheck, lint, tests, and production build — and report a concise pass/fail summary. Use before opening a PR, before recording the demo, or whenever the user asks to confirm the build is shippable.
---

# Quality gate

Run the project's checks in order and report results. Stop and surface the first failing stage clearly;
do not paper over failures.

## Steps

1. **Preconditions.** If `package.json` is missing, report that the project isn't scaffolded yet and
   stop. If `node_modules` is missing, run `npm install` first.

2. **Run, in order** (skip a stage only if its script genuinely doesn't exist, and say so):
   - `npm run typecheck` — strict TS, must be clean.
   - `npm run lint` — ESLint, zero errors.
   - `npm test` — unit tests for the critical flows (auth, browse, purchase, cart checkout, error
     mapping).
   - `npm run build` — production build must succeed.

3. **Report** a compact summary table: each stage → ✅ / ❌, with the key error lines for any failure
   and a concrete next step. Conclude with an overall `SHIPPABLE: YES/NO`.

## Notes
- This is the same typecheck the Stop hook enforces — running it here surfaces problems earlier and
  covers lint/test/build too.
- For demo-readiness also confirm: dev server proxies `/graphql` and `/pictures`, and the app handles a
  cold load against the running backend (login probe, empty store, image 404 fallback).
