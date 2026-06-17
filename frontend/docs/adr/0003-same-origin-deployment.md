# ADR-0003: Same-origin deployment behind the backend ingress

**Status:** Accepted · **Date:** 2026-06-17

## Context

The whole system must run locally on Docker + Minikube, and a grader must run it **without reading
source** (reading source to find how to run = failed submission). The backend already exposes an nginx
ingress on host `petstore.local` over self-signed TLS and **sends no CORS headers**
(`docs/BACKEND_INTEGRATION.md` §8). The frontend is a separate repo that must compose into the same
cluster.

## Decision

Serve the built SPA **same-origin behind the backend's existing nginx ingress**, path-routed:

- `/` → frontend static assets (nginx serving the Vite build)
- `/graphql` → backend API
- `/pictures/*` → backend image proxy

Everything is one origin (`https://petstore.local`): **one TLS cert to trust, no CORS, no preflight,
no mixed content.** In dev, the Vite dev server proxies `/graphql` and `/pictures` to the backend so the
same relative paths resolve at `localhost:5173`.

## Consequences

- The app calls **relative** `/graphql` and `/pictures` — no API base-URL config in production; the
  same code works in dev (via proxy) and in-cluster (via ingress).
- The frontend ships a **Dockerfile** (multi-stage: build → static nginx) and **Kubernetes manifests**
  (Deployment + Service + an ingress path rule, or a patch to the shared ingress).
- Composition across the two repos is a single documented `make`-style up command; the README is the
  grader's blind-run path. (Deployment phase — see `docs/PLAN.md`.)
- No CORS handling is needed anywhere. If the frontend were ever served from a different origin, the
  backend would have to add CORS + allow the `Authorization` header — avoided by this layout.

## Alternatives considered

- **Separate origin/port for the SPA + backend CORS.** Requires backend changes (it sends no CORS
  today), adds preflight latency, and means trusting/relating two origins. Rejected.
- **Bundling the SPA into the Go binary.** Couples the repos tightly and complicates independent
  frontend iteration. Rejected in favor of a dedicated frontend container behind the shared ingress.
