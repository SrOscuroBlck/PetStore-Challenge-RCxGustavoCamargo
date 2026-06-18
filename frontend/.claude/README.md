# Development framework (`.claude/`)

Project-specific harness for building the customer storefront. The goal: encode every non-negotiable
from `CLAUDE.md` as something **mechanical** — enforced by hooks, assisted by skills, audited by agents —
so quality doesn't depend on remembering.

## Hooks (automatic guardrails) — `hooks/` + wired in `settings.json`

| Hook | Fires on | What it does |
|---|---|---|
| `guard.js` | Pre Edit/Write/MultiEdit | Blocks edits that violate invariants: `localStorage`, `dangerouslySetInnerHTML`, merchant ops (`createPet`/`removePet`/`soldPets`/`unsoldPets`), breeder fields, `any` on code, `first > 100`. Each rule has an escape hatch (e.g. `// safe-localStorage`, `// allow-any`). |
| `bash-guard.js` | Pre Bash | Refuses to `git add` a real `.env` (secrets stay out of VCS). |
| `verify-changes.js` | Post Edit/Write on `.ts/.tsx` | Runs `eslint --fix` on the changed file; surfaces anything left. No-ops until scaffolded. |
| `quality-gate.js` | Stop | Won't let a turn finish with a failing `typecheck` when TS files changed. Self-gating + loop-safe. |

All hooks **no-op gracefully before the project is scaffolded** (no `package.json`/`node_modules`), so
they won't get in the way during Phase 1.

**Escape hatches** (use sparingly, only for genuine exceptions): add the marker comment on the same
line — `// safe-localStorage`, `// safe-html`, `// allow-any`.

## Skills (`/`-invokable workflows) — `skills/`

- **`new-operation`** — add a typed GraphQL op end-to-end (validate against schema → codegen → wire hook).
- **`new-component`** — scaffold a component with the required loading/empty/error states, a11y, and
  reduced-motion-aware motion.
- **`quality-gate`** — run typecheck + lint + test + build and report `SHIPPABLE: YES/NO`.

## Agents (on-demand reviewers) — `agents/`

- **`contract-checker`** — read-only; validates GraphQL/API code against `docs/schema.graphqls` and the
  customer-only contract.
- **`ux-state-auditor`** — read-only; audits a screen for designed async states, a11y, and motion.

## Tuning
Edit `settings.json` to disable a hook, or the rule list in `hooks/guard.js` to adjust patterns.
Personal overrides go in `settings.local.json` (gitignored).
