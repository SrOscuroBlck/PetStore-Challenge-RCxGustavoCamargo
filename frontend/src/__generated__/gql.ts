/* eslint-disable */
import * as types from './graphql';
import type { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';

/**
 * Map of all GraphQL operations in the project.
 *
 * This map has several performance disadvantages:
 * 1. It is not tree-shakeable, so it will include all operations in the project.
 * 2. It is not minifiable, so the string of a GraphQL query will be multiple times inside the bundle.
 * 3. It does not support dead code elimination, so it will add unused operations.
 *
 * Therefore it is highly recommended to use the babel or swc plugin for production.
 * Learn more about it here: https://the-guild.dev/graphql/codegen/plugins/presets/preset-client#reducing-bundle-size
 */
type Documents = {
    "query AvailablePets($storeId: ID!, $species: Species, $first: Int, $after: String) {\n  availablePets(\n    storeId: $storeId\n    species: $species\n    first: $first\n    after: $after\n  ) {\n    edges {\n      node {\n        id\n        name\n        species\n        ageYears\n        description\n        pictureUrl\n        status\n        createdAt\n      }\n      cursor\n    }\n    pageInfo {\n      hasNextPage\n      endCursor\n    }\n  }\n}": typeof types.AvailablePetsDocument,
    "mutation Checkout($petIds: [ID!]!) {\n  checkout(petIds: $petIds) {\n    id\n    status\n  }\n}": typeof types.CheckoutDocument,
    "mutation PurchasePet($petId: ID!) {\n  purchasePet(petId: $petId) {\n    id\n    status\n  }\n}": typeof types.PurchasePetDocument,
};
const documents: Documents = {
    "query AvailablePets($storeId: ID!, $species: Species, $first: Int, $after: String) {\n  availablePets(\n    storeId: $storeId\n    species: $species\n    first: $first\n    after: $after\n  ) {\n    edges {\n      node {\n        id\n        name\n        species\n        ageYears\n        description\n        pictureUrl\n        status\n        createdAt\n      }\n      cursor\n    }\n    pageInfo {\n      hasNextPage\n      endCursor\n    }\n  }\n}": types.AvailablePetsDocument,
    "mutation Checkout($petIds: [ID!]!) {\n  checkout(petIds: $petIds) {\n    id\n    status\n  }\n}": types.CheckoutDocument,
    "mutation PurchasePet($petId: ID!) {\n  purchasePet(petId: $petId) {\n    id\n    status\n  }\n}": types.PurchasePetDocument,
};

/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 *
 *
 * @example
 * ```ts
 * const query = graphql(`query GetUser($id: ID!) { user(id: $id) { name } }`);
 * ```
 *
 * The query argument is unknown!
 * Please regenerate the types.
 */
export function graphql(source: string): unknown;

/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "query AvailablePets($storeId: ID!, $species: Species, $first: Int, $after: String) {\n  availablePets(\n    storeId: $storeId\n    species: $species\n    first: $first\n    after: $after\n  ) {\n    edges {\n      node {\n        id\n        name\n        species\n        ageYears\n        description\n        pictureUrl\n        status\n        createdAt\n      }\n      cursor\n    }\n    pageInfo {\n      hasNextPage\n      endCursor\n    }\n  }\n}"): (typeof documents)["query AvailablePets($storeId: ID!, $species: Species, $first: Int, $after: String) {\n  availablePets(\n    storeId: $storeId\n    species: $species\n    first: $first\n    after: $after\n  ) {\n    edges {\n      node {\n        id\n        name\n        species\n        ageYears\n        description\n        pictureUrl\n        status\n        createdAt\n      }\n      cursor\n    }\n    pageInfo {\n      hasNextPage\n      endCursor\n    }\n  }\n}"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "mutation Checkout($petIds: [ID!]!) {\n  checkout(petIds: $petIds) {\n    id\n    status\n  }\n}"): (typeof documents)["mutation Checkout($petIds: [ID!]!) {\n  checkout(petIds: $petIds) {\n    id\n    status\n  }\n}"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "mutation PurchasePet($petId: ID!) {\n  purchasePet(petId: $petId) {\n    id\n    status\n  }\n}"): (typeof documents)["mutation PurchasePet($petId: ID!) {\n  purchasePet(petId: $petId) {\n    id\n    status\n  }\n}"];

export function graphql(source: string) {
  return (documents as any)[source] ?? {};
}

export type DocumentType<TDocumentNode extends DocumentNode<any, any>> = TDocumentNode extends DocumentNode<  infer TType,  any>  ? TType  : never;