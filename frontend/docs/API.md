# API Reference

GraphQL is served at `POST /graphql` over **TLS** (plaintext HTTP is refused) behind HTTP Basic auth (every request must carry valid credentials). Identity and store scope are derived from the authenticated principal — no operation accepts a `storeId` argument. Errors carry a stable machine-readable `code` in `extensions`:

| code | meaning |
|---|---|
| `VALIDATION` | an input field failed validation |
| `UNAUTHENTICATED` | no valid credentials |
| `FORBIDDEN` | wrong role for the operation |
| `NOT_FOUND` | pet does not exist in the caller's store |
| `CONFLICT` | pet is no longer in a state that allows the operation |
| `UNAVAILABLE` | pet is no longer available |
| `UNSUPPORTED_MEDIA_TYPE` | picture is not JPEG/PNG/WebP |
| `PAYLOAD_TOO_LARGE` | picture exceeds the size cap |
| `COMPLEXITY_LIMIT_EXCEEDED` | the query is too expensive (e.g. a `first` far above the page cap) |
| `GRAPHQL_VALIDATION_FAILED` | the query is malformed or selects unknown fields |
| `INTERNAL` | unexpected server error |

List queries are Relay cursor connections (`first`/`after` → `edges { node cursor }`, `pageInfo { hasNextPage endCursor }`), keyset-ordered. `pictureUrl` is a same-origin path (`/pictures/{objectKey}`) the API serves over the same TLS, streaming the image from object storage; clients use it directly as an image source and never see a signed URL or the storage bucket (see [ADR-0007](adr/0007-picture-proxy-path.md)).

**Hardening:** schema introspection is disabled outside development (`GRAPHQL_INTROSPECTION`), every query is bounded by a complexity limit, and picture uploads are size-capped.

---

## Merchant operations

All require the **merchant** role and are scoped to the merchant's own store.

### `createPet` (mutation)
Creates a listing. The picture is uploaded via multipart (`Upload`), validated for content type and size, and stored in object storage; only its key is persisted.

- **Arguments:** `input: CreatePetInput!` — `name`, `species` (`CAT|DOG|FROG`), `ageYears`, `picture: Upload!`, `description`, `breederName`, `breederEmail`.
- **Returns:** `Pet!` (includes `createdAt`, `status = AVAILABLE`, resolvable `pictureUrl`).
- **Errors:** `VALIDATION`, `UNSUPPORTED_MEDIA_TYPE`, `PAYLOAD_TOO_LARGE`, `UNAUTHENTICATED`, `FORBIDDEN`.

```graphql
mutation($input: CreatePetInput!) {
  createPet(input: $input) { id status createdAt pictureUrl }
}
# multipart: variable "input.picture" bound to an uploaded file
```

### `removePet` (mutation)
Removes an **available** pet from the store.

- **Arguments:** `id: ID!`
- **Returns:** `Pet!` (`status = REMOVED`).
- **Errors:** `NOT_FOUND` (absent or another store's pet), `CONFLICT` (already sold/removed), `VALIDATION`, `UNAUTHENTICATED`, `FORBIDDEN`.

```graphql
mutation { removePet(id: "…") { id status } }
```

### `soldPets` (query)
Pets sold within an inclusive `[from, to]` timestamp range, newest-keyset paginated by `sold_at`.

- **Arguments:** `from: Time!`, `to: Time!`, `first: Int` (default 20, max 100), `after: String`.
- **Returns:** `PetConnection!`.
- **Errors:** `VALIDATION` (bad `after` cursor), `UNAUTHENTICATED`, `FORBIDDEN`.

```graphql
query($from: Time!, $to: Time!) {
  soldPets(from: $from, to: $to, first: 20) {
    edges { node { id name soldAt } cursor }
    pageInfo { hasNextPage endCursor }
  }
}
```

### `unsoldPets` (query)
The store's not-yet-sold (`AVAILABLE`) pets, keyset paginated by `created_at`.

- **Arguments:** `first: Int` (default 20, max 100), `after: String`.
- **Returns:** `PetConnection!`.
- **Errors:** `VALIDATION` (bad `after` cursor), `UNAUTHENTICATED`, `FORBIDDEN`.

```graphql
query { unsoldPets(first: 20) { edges { node { id name } cursor } pageInfo { hasNextPage endCursor } } }
```

> The `Pet` type exposes breeder contact fields because every operation here is merchant-only. Customer operations use `PublicPet`, which omits breeder PII.

---

## Customer operations

All require the **customer** role. They return `PublicPet` — the same pet without breeder contact fields, so breeder PII is never exposed to customers. A purchase is race-safe: under concurrent attempts on the same pet, exactly one succeeds.

### `availablePets` (query)
A store's not-yet-sold pets, keyset paginated by `created_at` (oldest first). Customers are not store-scoped, so the store is a client argument; sold/removed pets never appear. An unknown `storeId` yields an empty connection.

- **Arguments:** `storeId: ID!`, `species: Species` (optional filter — `CAT|DOG|FROG`; omitted/null returns all species), `first: Int` (default 20, max 100), `after: String`.
- **Returns:** `PublicPetConnection!`.
- **Errors:** `VALIDATION` (bad `storeId`/`after`), `UNAUTHENTICATED`, `FORBIDDEN`.
- The species filter is applied server-side before pagination, so `first`/`after`/`pageInfo` reflect the filtered set; cursors are keyset `(created_at, id)` and stay consistent within a given filter value. A species with no available pets returns an empty connection (not an error).

```graphql
query($storeId: ID!, $species: Species) {
  availablePets(storeId: $storeId, species: $species, first: 20) {
    edges { node { id name species pictureUrl } cursor }
    pageInfo { hasNextPage endCursor }
  }
}
```

### `purchasePet` (mutation)
Instantly buys one available pet for the authenticated customer. Idempotent — re-purchasing a pet you already own succeeds.

- **Arguments:** `petId: ID!`
- **Returns:** `PublicPet!` (`status = SOLD`).
- **Errors:** `UNAVAILABLE` (sold to someone else / removed), `NOT_FOUND` (no such pet), `VALIDATION`, `UNAUTHENTICATED`, `FORBIDDEN`.

```graphql
mutation { purchasePet(petId: "…") { id status } }
```

### `checkout` (mutation)
Buys several pets atomically — all or nothing. If any pet is unavailable the whole checkout fails and the error message names every unavailable pet.

- **Arguments:** `petIds: [ID!]!`
- **Returns:** `[PublicPet!]!` (each `SOLD`).
- **Errors:** `UNAVAILABLE` — message lists the unavailable pets' names; plus `VALIDATION`, `UNAUTHENTICATED`, `FORBIDDEN`.

```graphql
mutation($petIds: [ID!]!) { checkout(petIds: $petIds) { id status } }
```
