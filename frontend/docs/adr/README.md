# Architecture Decision Records

Short, dated records of the consequential decisions on the customer storefront — the *why* behind
choices a future reader (or grader) would otherwise have to reverse-engineer. Each ADR states the
context, the decision, its consequences, and the alternatives weighed.

The backend repo uses the same convention (e.g. ADR-0007, the picture-proxy path). These cover the
frontend.

| # | Decision | Status |
|---|---|---|
| [0001](0001-frontend-stack-and-graphql-client.md) | Frontend stack & GraphQL client (Apollo over urql) | Accepted |
| [0002](0002-authentication-and-session.md) | HTTP Basic auth, login-by-probe, sessionStorage credential | Superseded by 0005 |
| [0003](0003-same-origin-deployment.md) | Serve the SPA same-origin behind the backend ingress | Accepted |
| [0004](0004-catalog-freshness-and-optimistic-concurrency.md) | Catalog freshness & optimistic concurrency | Accepted |
| [0005](0005-open-storefront-gateway-auth.md) | Open storefront — credential injected at the gateway, no login wall | Accepted (amended by 0006) |
| [0006](0006-login-to-place-orders.md) | Login to place orders (open browse, gated ordering) | Accepted |
