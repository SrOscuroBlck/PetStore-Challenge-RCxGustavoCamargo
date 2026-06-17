# Backend integration guide (for the customer frontend)

How the React/TypeScript customer app talks to the Go GraphQL backend. Pair this with
`docs/API.md` (operation contract), `docs/schema.graphqls` (codegen source),
`docs/SECURITY.md` (security model), and `docs/challenge.md` (company requirements).

> The backend adapts to nothing here — the frontend conforms to this contract.

---

## 1. Transport & base URL

- GraphQL is **`POST /graphql` over HTTPS only**. Plaintext HTTP is refused.
- Unauthenticated health probe: `GET /healthz` → `{"status":"ok"}`.
- **Reach it two ways locally:**
  - Port-forward (dev): `kubectl port-forward -n petstore svc/petstore-api 8443:8443` →
    `https://localhost:8443/graphql`.
  - Ingress (recommended for the demo): nginx ingress on host `petstore.local`, TLS.
- **TLS is self-signed.** A browser must trust the cert once. The clean answer is to serve the
  frontend **same-origin** behind the shared ingress (see §8) so there is exactly one cert to accept
  and **no CORS**.

---

## 2. Authentication — HTTP Basic, no tokens

There is **no login/signup mutation and no session/JWT.** Auth is HTTP Basic on every request.

- Header: `Authorization: Basic ` + `base64("<email>:<password>")`.
  ```ts
  const authHeader = `Basic ${btoa(`${email}:${password}`)}`; // emails are ASCII; btoa is fine
  ```
- Attach it to **every** `/graphql` request (Apollo `setContext`/authLink, or urql `fetchOptions`).
- **"Login" flow:** collect email + password, then validate by firing one cheap authed query
  (e.g. `availablePets` for the current store). `UNAUTHENTICATED` / HTTP 401 → wrong credentials.
- **Credential storage:** keep the Basic value **in memory** (React state/context) or `sessionStorage`
  at most. **Never `localStorage`** (XSS exfiltration risk). Clear it on logout.
- **Demo customer:** `customer@petstore.local` / `demo-password`.
- **This app is customer-only.** Merchant operations return `FORBIDDEN`; never call or surface them.

---

## 3. Store identity — comes from the URL

- `availablePets` requires a `storeId: ID!`. There is **no store-list/discovery query.**
- Carry the store in the route, e.g. `/store/:storeId`, and pass it through to `availablePets`.
- The demo store id is **fixed**: `11111111-1111-1111-1111-111111111111` (the seeder pins it; `make k8s-up` also prints it). Open `/store/11111111-1111-1111-1111-111111111111`.
- An unknown `storeId` yields an **empty connection** (not an error) — render a friendly empty state.

---

## 4. Customer operations

Customer queries return **`PublicPet`** — it has **no breeder fields** (breeder PII is never exposed;
those fields don't exist on the type, so selecting them is a hard validation error).

`PublicPet` fields: `id`, `name`, `species` (`CAT|DOG|FROG`), `ageYears`, `description`, `pictureUrl`,
`status` (`AVAILABLE|SOLD|REMOVED`), `createdAt`, `soldAt`.

| Operation | Type | Signature | Returns | Notes |
|---|---|---|---|---|
| `availablePets` | query | `(storeId: ID!, species: Species, first: Int, after: String)` | `PublicPetConnection!` | Only `AVAILABLE` pets; optional `species` (`CAT\|DOG\|FROG`) filters server-side before pagination. |
| `purchasePet` | mutation | `(petId: ID!)` | `PublicPet!` (`SOLD`) | Idempotent for the same buyer. |
| `checkout` | mutation | `(petIds: [ID!]!)` | `[PublicPet!]!` (each `SOLD`) | All-or-nothing. |

```graphql
query AvailablePets($storeId: ID!, $species: Species, $first: Int, $after: String) {
  availablePets(storeId: $storeId, species: $species, first: $first, after: $after) {
    edges { node { id name species ageYears description pictureUrl status createdAt } cursor }
    pageInfo { hasNextPage endCursor }
  }
}
mutation PurchasePet($petId: ID!) { purchasePet(petId: $petId) { id status } }
mutation Checkout($petIds: [ID!]!) { checkout(petIds: $petIds) { id status } }
```

Generate typed hooks for these from `docs/schema.graphqls` with graphql-codegen — don't hand-type them.

---

## 5. Pagination — Relay cursor connections

- `first` (default 20, **max 100**) + `after` (opaque cursor). Read `pageInfo.hasNextPage` /
  `pageInfo.endCursor` and pass `endCursor` as the next `after`.
- Build the catalog as **infinite scroll / load-more**; never request more than 100 (you'll get
  `COMPLEXITY_LIMIT_EXCEEDED`). This is how the <2s / 1k-concurrent target is met — paginate, don't
  bulk-load.

---

## 6. Errors → UX mapping

GraphQL errors arrive as `errors[]`, each with a stable `extensions.code` and a human-readable
`message`. Surface the `message` for the user-facing cases; branch on `code` for behavior.

| `code` | What it means | Suggested UX |
|---|---|---|
| `UNAVAILABLE` | pet already sold/removed | Toast the message; refresh the card to SOLD. **For `checkout`, the message names every unavailable pet — show those names** (challenge requirement). |
| `NOT_FOUND` | no such pet in scope | "This pet is no longer listed." |
| `UNAUTHENTICATED` | bad/missing creds | Send back to the login screen. |
| `FORBIDDEN` | wrong role | Shouldn't happen in a customer app; log + generic error. |
| `VALIDATION` | bad input (e.g. malformed id) | Generic "something went wrong"; this is a client bug. |
| `COMPLEXITY_LIMIT_EXCEEDED` | `first` too large | Don't exceed 100 — fix the query. |
| `GRAPHQL_VALIDATION_FAILED` | malformed query / unknown field | Build-time bug; fix the operation. |
| `INTERNAL` | server error | Generic, friendly fallback; don't leak details. |

The error messages are already written to be human-readable — prefer showing them over inventing your own,
especially the checkout one that lists pet names.

---

## 7. Race conditions & freshness

- Availability is **server-derived** — two customers can race the same pet; exactly one wins.
- Use **optimistic UI** for purchase/checkout, but **roll back and show the `UNAVAILABLE` message** when
  the server rejects. Never assume a locally-shown pet is still buyable.
- A page refresh must reflect sold/removed pets (don't render `SOLD`/`REMOVED` in the catalog).

---

## 8. Same-origin deployment (recommended) & CORS

- The backend sends **no CORS headers** today. Avoid the problem entirely by serving the SPA
  **same-origin** behind the existing nginx ingress (`petstore.local`): path-route `/graphql` → backend,
  `/` → the frontend. One TLS cert, no CORS, no preflight.
- If you ever serve the frontend from a different origin/port, the backend must add CORS (allow your
  origin + the `Authorization` header). The same-origin layout avoids this; prefer it.

---

## 9. Pet images

- `pictureUrl` is a **same-origin path** the API serves: `/pictures/{objectKey}` (e.g.
  `/pictures/pets/<uuid>`). Use it directly as an `<img>` `src`.
- The API streams the image over its own TLS endpoint (no presigned URL, no separate object-storage
  host, no `http`/`https` mixed content). It works over the same `kubectl port-forward` you use for
  `/graphql`, and same-origin behind the ingress in the deployed setup.
- **Dev proxy:** when the SPA runs on a different origin (e.g. Vite at `:5173`), proxy `/pictures` to
  the backend exactly as you proxy `/graphql`, so the relative `pictureUrl` resolves.
- The path is unauthenticated (public catalog images, opaque keys) and sends
  `Cache-Control: public, max-age=300`; render a graceful placeholder on a `404` (removed pet / unknown key).

---

## 10. Quick curl reference

```bash
# Health
curl -k https://localhost:8443/healthz

# Validate login (customer) — 200 with data = good creds, 401 = bad
curl -k -u customer@petstore.local:demo-password -H 'Content-Type: application/json' \
  -d "{\"query\":\"{ availablePets(storeId:\\\"$STORE\\\", first:1){ edges{ node{ id } } } }\"}" \
  https://localhost:8443/graphql

# Purchase
curl -k -u customer@petstore.local:demo-password -H 'Content-Type: application/json' \
  -d "{\"query\":\"mutation{ purchasePet(petId:\\\"$PET\\\"){ id status } }\"}" \
  https://localhost:8443/graphql
```
