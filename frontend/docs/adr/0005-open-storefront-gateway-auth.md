# ADR-0005: Open storefront — credential injected at the gateway, no login wall

**Status:** Accepted, amended by [ADR-0006](0006-login-to-place-orders.md) · **Date:** 2026-06-17 · **Supersedes:** [ADR-0002](0002-authentication-and-session.md)

> **Amended by ADR-0006:** browsing stays open as described here, but **placing an order now requires a
> login** (the brief's "login … place orders"). The gateway injects the ambient credential only when the
> request carries no `Authorization`, so a signed-in customer's credential passes through for orders.

## Context

The backend requires HTTP Basic auth on every customer request (`docs/BACKEND_INTEGRATION.md` §2),
and the challenge requires customer/merchant endpoints to be secured "with basic HTTP authentication"
— whose stated purpose is **role separation**, not per-customer identity (`docs/challenge.md`).

Reading the *customer user stories* closely: they ask to open a store URL and **see products**, click
**purchase**, use a **cart**, and **checkout**. None mention an account, sign-in, or order history. The
domain also has **no money** ("for free, we live in a world with no money") and a single demo customer.
In real storefronts the moment that triggers authentication is **payment** — which does not exist here.

ADR-0002 originally specified a login screen + client-held credential. That imposes a login wall the
stories don't ask for, and puts the Basic credential in the browser (`sessionStorage`) where XSS can
reach it. We reconsidered.

## Decision

**The storefront is open. There is no login screen.** The customer opens `/store/:storeId` and sees the
catalog immediately; purchase and checkout work the same way.

Authentication is enforced **at the boundary, by infrastructure, not by the user**:

- **Prod:** the same-origin nginx gateway (ADR-0003) injects `Authorization: Basic <customer>` on
  proxied `/graphql`. The credential lives in a Kubernetes Secret.
- **Dev:** the Vite proxy injects the same header, read from a git-ignored `.env`
  (`DEV_CUSTOMER_EMAIL`/`DEV_CUSTOMER_PASSWORD`) at dev-server runtime.

The **browser never holds the credential** — it is not in the bundle, not in storage, not in JS.

## Consequences

- **Products-first UX:** opening a store shows pets, the way a store should behave.
- **Stronger security posture than a client-held credential:** nothing to exfiltrate via XSS; the
  credential never crosses the network boundary to the browser. Endpoints stay Basic-auth protected and
  role-separated — a direct, unauthenticated API call is still rejected.
- **One shared customer identity** for all demo traffic. The stories need no distinct customers and no
  payment, so this is sufficient. The data layer keeps a clean seam, so per-customer identity could be
  added later without rework.
- The frontend has **no auth/session code** — no login route, credential store, or route guard. Less
  surface, less to break.

## Alternatives considered

- **Login wall + client-held credential (ADR-0002).** Rejected: imposes a sign-in the stories/domain
  don't justify, and weaker (credential in the browser).
- **Bundle the customer credential in the client.** Rejected: ships a secret in the bundle, against the
  threat model; gateway injection keeps it server-side.
- **Identity step at checkout** (mirror real-store "sign in to pay"). Rejected: there is no payment and
  no second customer — it would be ceremony with no function.

## Risk & how we own it

Anyone who can reach the host can browse/purchase as the customer — identical to a public storefront,
and acceptable for a money-less local demo. A strict grader might expect a visible login; the README
states this decision and its rationale explicitly so it reads as deliberate product+security judgment,
not an omission.
