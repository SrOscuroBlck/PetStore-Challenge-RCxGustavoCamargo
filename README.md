# Pet Store — Backend

A backend for a multi-tenant pet store. **Merchants** manage their store's pet listings through a GraphQL API; **customers** browse available pets and purchase them individually or via a cart checkout. Built for the Robotic Crew challenge (`docs/challenge.md`).

This repository is the **backend only** and runs entirely on your machine — no externally-hosted services.

> **Status:** early development. This README documents what has been **decided** (stack and architecture). Operational sections — how to run it, configure it, and call the API — are added to this file and to [`docs/API.md`](docs/API.md) **as each piece is actually built**, so the documentation always reflects the code rather than predicting it.

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

## Run on Minikube (full stack)

Runs the entire system — Postgres, Redis, MinIO, database migrations, and the API — on local
Kubernetes via Minikube. Each component is its own Deployment + Service; the API talks to them
over in-cluster DNS. Requires **Docker** (running), **minikube**, **kubectl**, **openssl**, and Go 1.25.

```bash
make k8s-up      # one command: start minikube, build the image, apply manifests,
                 # run migrations, and seed demo accounts
```

`make k8s-up` starts Minikube (if needed), enables the ingress addon, builds the API image into the
cluster, applies the manifests in [`deploy/k8s/`](deploy/k8s/), creates the Secrets and the
migrations ConfigMap (secret values are generated at deploy time, never committed), brings up
Postgres/Redis/MinIO, then rolls out the API. **Migrations run automatically as an init container
before the API serves traffic.** A one-shot Job then seeds demo accounts:

| Role | Email | Password |
|---|---|---|
| Merchant (owns a "Demo Store") | `merchant@petstore.local` | `demo-password` |
| Customer | `customer@petstore.local` | `demo-password` |

**Reach the API over TLS** from your host (the API serves HTTPS only):

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

**Explore in a browser (GraphQL playground).** With the same `port-forward` running, open
**`https://localhost:8443/playground`** and accept the self-signed-cert warning. The page is a
GraphiQL editor with full schema docs and autocompletion (it is served only when
`GRAPHQL_INTROSPECTION=true`, which the Minikube config sets; it is off by default in production).
The page loads without credentials, but everything it sends — including the introspection query that
fills the schema docs — is behind Basic auth, so set a header first or the docs panel stays empty.
Open the **Headers** pane at the bottom and paste one of:

```json
{ "Authorization": "Basic bWVyY2hhbnRAcGV0c3RvcmUubG9jYWw6ZGVtby1wYXNzd29yZA==" }
```
```json
{ "Authorization": "Basic Y3VzdG9tZXJAcGV0c3RvcmUubG9jYWw6ZGVtby1wYXNzd29yZA==" }
```

The first is the merchant, the second the customer (`base64("<email>:demo-password")`); switch headers
to switch roles. From there you can run `unsoldPets`, `soldPets`, `availablePets`, `purchasePet`, and
`checkout`. One exception: **`createPet` uploads a file**, which a browser GraphiQL can't send — create
the first pet with the `curl` multipart command (see the merchant step in the demo flow), then browse
and purchase it from the playground.

```bash
make logs        # tail the API logs
make k8s-down    # tear the stack down (removes the petstore namespace)
```

**Ingress.** An nginx Ingress (TLS, host `petstore.local`, re-encrypting to the HTTPS backend) is
also applied. On Linux, add `petstore.local` to `/etc/hosts` pointing at `minikube ip` and
`curl -k https://petstore.local/healthz`. On macOS with the Docker driver the node IP isn't routable
from the host, so reach the ingress with `minikube tunnel` (separate terminal, needs sudo) or simply
use the `kubectl port-forward` command above — it is the supported default.

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
