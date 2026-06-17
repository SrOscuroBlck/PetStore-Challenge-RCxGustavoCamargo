# Pet Store — Customer Storefront

The customer-facing web frontend for the Robotic Crew pet-store challenge. A customer opens a merchant
store's URL, browses the pets available there, and buys them — individually or via a cart checkout. The
merchant experience is API-only by design and has no frontend here.

> **Scope:** customer user stories only. The Go GraphQL backend is a separate, finished repo; this app
> conforms to its fixed contract.

## Status

Early development. The development framework, architecture, and documentation are in place; application
code is built in phases — see **[docs/PLAN.md](docs/PLAN.md)**. Run instructions below are filled in as
the corresponding phases land (the full one-command local run arrives in the deployment phase).

## Tech stack

React + TypeScript (strict) · Vite · React Router · Apollo Client + graphql-codegen · Tailwind CSS +
Framer Motion. Rationale and trade-offs are recorded as ADRs — see **[docs/adr/](docs/adr/)**.

## How it talks to the backend

GraphQL only, `POST /graphql` over TLS. Customer endpoints require **HTTP Basic auth**. **Browsing is
open** — the same-origin gateway (nginx in prod, the Vite proxy in dev) injects an ambient credential, so
the catalog loads with no login. **Placing an order requires sign-in** — Buy/Checkout open a login
dialog; the signed-in customer's credential is attached client-side and the gateway passes it through.
This is a deliberate product + security decision — see
**[ADR-0006](docs/adr/0006-login-to-place-orders.md)** (amending
[ADR-0005](docs/adr/0005-open-storefront-gateway-auth.md)). Demo customer:
`customer@petstore.local` / `demo-password`. Store identity comes from the URL (`/store/:storeId`); pet
images load from the API's same-origin `/pictures/{key}` path. Full contract:
**[docs/API.md](docs/API.md)**, **[docs/BACKEND_INTEGRATION.md](docs/BACKEND_INTEGRATION.md)**,
**[docs/schema.graphqls](docs/schema.graphqls)**.

The app is served **same-origin behind the backend's nginx ingress** (`petstore.local`), so there is one
TLS cert and no CORS — see **[ADR-0003](docs/adr/0003-same-origin-deployment.md)**.

## Local development

> Requires the backend running locally (Docker + Minikube) per the backend repo's README, and Node 20+.

```bash
cp .env.example .env          # set the proxy target + the dev customer credential (DEV_CUSTOMER_*)
npm install
npm run codegen               # generate typed hooks from docs/schema.graphqls
npm run dev                   # Vite dev server; proxies /graphql (+ injected auth) and /pictures
```

Then open `http://localhost:5173/store/<storeId>` — the catalog loads with no login step (the dev proxy
authenticates requests). Get the demo store id from the running cluster
(see [docs/BACKEND_INTEGRATION.md](docs/BACKEND_INTEGRATION.md) §3). The dev credential defaults to the
demo customer `customer@petstore.local` / `demo-password`.

### Scripts

| Script | Purpose |
|---|---|
| `npm run dev` | Dev server with HMR + backend proxy |
| `npm run codegen` | Regenerate GraphQL types/hooks from the schema (run after editing any operation) |
| `npm run build` / `npm run preview` | Production build / preview it |
| `npm run typecheck` / `npm run lint` | Strict TS check / ESLint |
| `npm test` | Unit tests for the critical flows |

*(Scripts are wired up in the scaffold phase; this table is the intended interface.)*

## Running the full system (Docker + Minikube)

The frontend deploys into the same local cluster as the backend, served same-origin behind the shared
ingress. The single documented `make`-style up command and the grader-blind walkthrough are finalized in
the deployment phase (see [docs/PLAN.md](docs/PLAN.md), Phase 7).

## Documentation

| Doc | What it covers |
|---|---|
| [docs/challenge.md](docs/challenge.md) | The company's requirements — source of truth |
| [docs/PLAN.md](docs/PLAN.md) | Build phases, user-story acceptance criteria, status |
| [docs/DESIGN.md](docs/DESIGN.md) | Visual design system — tokens, type, motifs, components |
| [docs/adr/](docs/adr/) | Architecture decision records |
| [docs/THREAT_MODEL.md](docs/THREAT_MODEL.md) | Frontend security model & mitigations |
| [docs/API.md](docs/API.md) · [docs/BACKEND_INTEGRATION.md](docs/BACKEND_INTEGRATION.md) · [docs/SECURITY.md](docs/SECURITY.md) | Backend contract & security (fixed) |

## Confidentiality

This challenge is confidential. Do not push to a public repository.
