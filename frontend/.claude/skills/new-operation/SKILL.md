---
name: new-operation
description: Add or change a typed GraphQL operation for the customer storefront end-to-end — write/validate the .graphql document against docs/schema.graphqls, run graphql-codegen, and wire the generated typed hook. Use whenever adding or editing an availablePets / purchasePet / checkout query or mutation. Guarantees no hand-written, untyped queries.
---

# Add a typed GraphQL operation

The contract is fixed (`docs/schema.graphqls`, `docs/API.md`). Every operation is typed via
graphql-codegen — never hand-write types or inline untyped `gql`.

## Steps

1. **Confirm the operation is allowed.** Customer-only: `availablePets` (query), `purchasePet`,
   `checkout` (mutations). If the request implies a merchant op (`createPet`, `removePet`, `soldPets`,
   `unsoldPets`) or breeder fields, stop — out of scope, would return FORBIDDEN / fail validation.

2. **Write the `.graphql` document** in the operations location (e.g. `src/graphql/operations/`).
   - Select only fields that exist on the target type. Customer queries return **`PublicPet`**
     (`id, name, species, ageYears, description, pictureUrl, status, createdAt, soldAt`) — no breeder
     fields.
   - For lists: full Relay shape — `edges { node { … } cursor }` and
     `pageInfo { hasNextPage endCursor }`. Variables `$storeId: ID!, $first: Int, $after: String`.
     Keep `first` ≤ 100.
   - Name operations clearly (e.g. `AvailablePets`, `PurchasePet`, `Checkout`).

3. **Run codegen:** `npm run codegen`. Fix any schema errors it reports (unknown field/type/arg means
   the document is wrong — the schema is right).

4. **Consume the generated hook** in components — never re-type the response. For pagination use the
   Apollo `fetchMore` + `relayStylePagination` cache policy. For mutations, add `optimisticResponse`
   and an `update`/rollback path; handle `UNAVAILABLE` (and for checkout, surface the message naming
   unavailable pets) by branching on `extensions.code`.

5. **Verify:** `npm run typecheck` passes; consider invoking the `contract-checker` agent on the new
   document for an independent contract review.

## Guardrails
- No `any` on the result. No inline untyped queries. No `first > 100`.
- If `docs/schema.graphqls` doesn't contain a field you need, the feature is out of contract — raise it,
  don't work around it.
