# Pet Store

A multi-tenant pet store. **Merchants** manage their store's pet listings through a GraphQL API; **customers** browse available pets in a web storefront and purchase them individually or via a cart checkout. Built for the Robotic Crew challenge (`docs/challenge.md`).

This is a **monorepo** and runs entirely on your machine — no externally-hosted services:

- **Backend** (this directory) — Go GraphQL API, Postgres, Redis, MinIO. Code under `cmd/`, `internal/`, `db/`.
- **Frontend** — React + TypeScript customer storefront in [`frontend/`](frontend/) (see [`frontend/README.md`](frontend/README.md)).
- **Deploy** — one command, [`make k8s-up`](#run-the-full-stack-on-minikube), brings the whole system up on local Kubernetes (Minikube) behind a same-origin gateway.

> The frontend is served by a same-origin nginx gateway that proxies `/graphql` and `/pictures` to the API. Browsing is open — the gateway injects an ambient customer credential, so anonymous browsing holds no secret in the browser; placing an order requires sign-in, and that signed-in credential lives in `sessionStorage` for the order session only (see [`frontend/docs/adr/`](frontend/docs/adr/)).

---

## Tech stack

| Concern | Choice |
|---|---|
| Language | Go 1.25 |
| API | GraphQL (`gqlgen`, schema-first) |
| Database | PostgreSQL (`pgx/v5`) with type-safe queries via `sqlc` |
| Schema & migrations | Atlas (versioned, linted) |
| Cache | Redis (`go-redis`) |
| Object storage | MinIO (self-hosted, S3-compatible) for pet pictures |
| Auth | HTTP Basic + bcrypt password hashing |
| Orchestration | Docker + Minikube (local Kubernetes) |

The reasoning behind each choice is recorded in [`docs/adr/`](docs/adr/).

---

## Architecture

The API is layered so business rules never depend on infrastructure: a pure domain core, application services that own use cases and transaction boundaries, and adapters that implement persistence, caching, and object storage behind interfaces. Multi-tenant isolation, race-safe purchasing, cached reads, and encryption in transit and at rest are designed in from the start.

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for the full picture.

---

## Development

Requires Go 1.25+ and `make`.

```bash
make tools                 # install gqlgen, sqlc, atlas, goimports, golangci-lint into ./bin
make check                 # format check, go vet, golangci-lint, build, race tests
```

Local dependencies and database (requires Docker):

```bash
cp .env.example .env   # then set PII_ENCRYPTION_KEY (see below); the Makefile auto-loads .env
make dev               # start Postgres, Redis, MinIO via docker-compose
make migrate-up        # apply the database schema
make tls-certs         # generate a local self-signed TLS cert into ./certs
make run               # run the server over TLS; GET /healthz returns {"status":"ok"}
```

`make run` serves **HTTPS only** (plaintext is refused) and connects to Postgres, MinIO, and Redis,
reading `PII_ENCRYPTION_KEY`, `REDIS_ADDR`, the `MINIO_*` vars, and `TLS_CERT_FILE`/`TLS_KEY_FILE`,
so it needs `make dev`, `make tls-certs`, and a populated `.env`. Generate the encryption key once
with `openssl rand -base64 32`; the MinIO bucket is created automatically at startup, and the
catalog cache falls back to Postgres if Redis is unavailable. Hit it with a self-signed cert via
`curl -k https://localhost:8443/healthz`. The GraphQL endpoint at `/graphql` is behind HTTP Basic
auth (requests without valid credentials get 401); schema introspection is off unless
`GRAPHQL_INTROSPECTION=true`, which also serves a browser GraphQL playground at
`https://localhost:8443/playground` (see the Minikube section for how to set the auth header).

Configuration is read from environment variables and validated at startup — a missing
required value aborts with an error naming it. See [`.env.example`](.env.example) for the
variables currently in use (more are added as features that need them land).

---

## Run the full stack on Minikube

Runs the **entire system** — Postgres, Redis, MinIO, database migrations, the Go API, and the React
storefront — on local Kubernetes via Minikube. Each component is its own Deployment + Service.
Requires **Docker** (running), **minikube**, **kubectl**, **openssl**, **Node 20**, and Go 1.25.

```bash
make k8s-up      # one command: start minikube, build the API + web images, apply manifests,
                 # run migrations, seed a demo catalog, and roll out the storefront
```

It builds both images into the cluster, applies the manifests in [`deploy/k8s/`](deploy/k8s/), creates
the Secrets and migrations ConfigMap (secret values generated at deploy time, never committed), brings
up Postgres/Redis/MinIO, rolls out the API (**migrations run as an init container before it serves**),
seeds demo data, then rolls out the web storefront behind the gateway. When it finishes it prints the
storefront URL. A one-shot Job seeds demo accounts and a catalog of pets:

| Role | Email | Password |
|---|---|---|
| Merchant (owns "Demo Store") | `merchant@petstore.local` | `demo-password` |
| Merchant (owns "Second Store") | `merchant2@petstore.local` | `demo-password` |
| Customer | `customer@petstore.local` | `demo-password` |
| Customer (second shopper) | `customer2@petstore.local` | `demo-password` |

The **second customer** lets you exercise the purchase/checkout race (buy a pet as one shopper and
watch the other get a human-readable `UNAVAILABLE` error). The **second merchant** owns a separate
store (`22222222-2222-2222-2222-222222222222`, stocked with its own pets) so you can demonstrate
store isolation: merchant 2 sees only its own pets and gets `NOT_FOUND` on merchant 1's.


The store is pre-filled with a catalog of cats, dogs, and frogs (with real bundled photos —
see [`cmd/seed/assets/CREDITS.md`](cmd/seed/assets/CREDITS.md)), so the customer storefront is
browsable immediately. The demo store id is fixed —
**`11111111-1111-1111-1111-111111111111`** — so the customer site opens at
`/store/11111111-1111-1111-1111-111111111111`. `make k8s-up` also prints it at the end. Seeding is
idempotent: re-running leaves existing demo data untouched.

**Open the customer storefront** — port-forward the web gateway and browse to the demo store:

```bash
kubectl port-forward -n petstore svc/petstore-web 8080:80     # leave running in one terminal
# then open http://localhost:8080/store/11111111-1111-1111-1111-111111111111
```

The gateway serves the SPA and proxies `/graphql` + `/pictures` to the API on the same origin,
injecting an ambient credential — so browsing works with no login and no secret in the browser
(placing an order prompts sign-in; that credential lives in `sessionStorage` for the order session).
(Via the ingress instead: add `petstore.local` to `/etc/hosts` → `minikube ip`, then browse
`https://petstore.local/store/11111111-1111-1111-1111-111111111111`; on the macOS Docker driver use
`minikube tunnel`.)

**Reach the API directly over TLS** (merchant operations, curl, the playground) — the API serves HTTPS only:

```bash
kubectl port-forward -n petstore svc/petstore-api 8443:8443   # leave running in one terminal
```

```bash
# Health (unauthenticated)
curl -k https://localhost:8443/healthz                        # {"status":"ok"}

# GraphQL behind Basic auth — a merchant lists their unsold pets
curl -k -u merchant@petstore.local:demo-password \
     -H 'Content-Type: application/json' \
     -d '{"query":"{ unsoldPets(first:5){ edges{ node{ id name } } } }"}' \
     https://localhost:8443/graphql
```

Plaintext HTTP is refused, and `/graphql` without credentials returns `401`.

Pet pictures are served by the API over the same TLS endpoint at `/pictures/{objectKey}` — a pet's
`pictureUrl` is that same-origin path, so a browser (or the customer frontend) loads images straight
from the API with no extra setup or separate object-storage exposure ([ADR-0007](docs/adr/0007-picture-proxy-path.md)).

**Explore in a browser (GraphQL playground).** With the same `port-forward` running, open
**`https://localhost:8443/playground`** and accept the self-signed-cert warning. The page is an
[Altair](https://altairgraphql.dev/) playground with full schema docs, autocompletion, and file
uploads (it is served only when `GRAPHQL_INTROSPECTION=true`, which the Minikube config sets; it is
off by default in production). The page loads without credentials, but everything it sends —
including the introspection query that fills the schema docs — is behind Basic auth, so set a header
first or the docs panel stays empty. Open the **Headers** panel and add `Authorization` with one of:

```
Basic bWVyY2hhbnRAcGV0c3RvcmUubG9jYWw6ZGVtby1wYXNzd29yZA==
```
```
Basic Y3VzdG9tZXJAcGV0c3RvcmUubG9jYWw6ZGVtby1wYXNzd29yZA==
```

The first is the merchant, the second the customer (`base64("<email>:demo-password")`); switch headers
to switch roles. **Every** operation runs here — `createPet`, `removePet` (merchant) and `availablePets`,
`purchasePet`, `checkout` (customer) — including the `createPet` picture upload: write the mutation with
an `$picture: Upload!` variable, then use Altair's **Add file** control (under Variables) to bind a file
to that variable.

```bash
make logs        # tail the API logs
make k8s-down    # tear the stack down (removes the petstore namespace)
```

**Ingress.** An nginx Ingress (TLS, host `petstore.local`) terminates TLS and routes to the web
gateway, which serves the storefront and proxies `/graphql` + `/pictures` to the API — so the whole
system is same-origin. On Linux, add `petstore.local` to `/etc/hosts` pointing at `minikube ip` and
browse `https://petstore.local`. On macOS with the Docker driver the node IP isn't routable from the
host, so reach the ingress with `minikube tunnel` (separate terminal, needs sudo) — or use the
`kubectl port-forward svc/petstore-web 8080:80` command above, which is the supported default.

---

## Documentation

| Document | What's inside |
|---|---|
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | Layering, request lifecycle, concurrency & race-condition strategy |
| [`docs/DATA_MODEL.md`](docs/DATA_MODEL.md) | Entity-relationship diagram, tables, indexes, what's encrypted |
| [`docs/SECURITY.md`](docs/SECURITY.md) | Authentication, store isolation, encryption, hardening, TLS |
| [`docs/API.md`](docs/API.md) | GraphQL operations reference — populated as each operation is implemented |
| [`docs/adr/`](docs/adr/) | Architecture Decision Records (the "why" behind key choices) |
| [`docs/challenge.md`](docs/challenge.md) | The original challenge brief |
