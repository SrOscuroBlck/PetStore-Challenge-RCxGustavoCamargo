/* eslint-disable */
import type { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  Time: { input: string; output: string; }
  Upload: { input: File; output: File; }
};

export type CreatePetInput = {
  ageYears: Scalars['Int']['input'];
  breederEmail: Scalars['String']['input'];
  breederName: Scalars['String']['input'];
  description: Scalars['String']['input'];
  name: Scalars['String']['input'];
  picture: Scalars['Upload']['input'];
  species: Species;
};

export type Mutation = {
  __typename?: 'Mutation';
  checkout: Array<PublicPet>;
  createPet: Pet;
  purchasePet: PublicPet;
  removePet: Pet;
};


export type MutationCheckoutArgs = {
  petIds: Array<Scalars['ID']['input']>;
};


export type MutationCreatePetArgs = {
  input: CreatePetInput;
};


export type MutationPurchasePetArgs = {
  petId: Scalars['ID']['input'];
};


export type MutationRemovePetArgs = {
  id: Scalars['ID']['input'];
};

export type PageInfo = {
  __typename?: 'PageInfo';
  endCursor?: Maybe<Scalars['String']['output']>;
  hasNextPage: Scalars['Boolean']['output'];
};

/**
 * Pet is the merchant-facing view of a pet listing. It includes breeder contact
 * details, which are visible only to the owning merchant; the customer-facing
 * catalog in a later issue must expose a separate type without these fields.
 */
export type Pet = {
  __typename?: 'Pet';
  ageYears: Scalars['Int']['output'];
  breederEmail: Scalars['String']['output'];
  breederName: Scalars['String']['output'];
  createdAt: Scalars['Time']['output'];
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  pictureUrl: Scalars['String']['output'];
  soldAt?: Maybe<Scalars['Time']['output']>;
  species: Species;
  status: PetStatus;
};

export type PetConnection = {
  __typename?: 'PetConnection';
  edges: Array<PetEdge>;
  pageInfo: PageInfo;
};

export type PetEdge = {
  __typename?: 'PetEdge';
  cursor: Scalars['String']['output'];
  node: Pet;
};

export enum PetStatus {
  Available = 'AVAILABLE',
  Removed = 'REMOVED',
  Sold = 'SOLD'
}

/**
 * PublicPet is the customer-facing view of a pet. It omits breeder contact details
 * so breeder PII is never exposed to customers, even though the cached domain pet
 * still carries them.
 */
export type PublicPet = {
  __typename?: 'PublicPet';
  ageYears: Scalars['Int']['output'];
  createdAt: Scalars['Time']['output'];
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  pictureUrl: Scalars['String']['output'];
  soldAt?: Maybe<Scalars['Time']['output']>;
  species: Species;
  status: PetStatus;
};

export type PublicPetConnection = {
  __typename?: 'PublicPetConnection';
  edges: Array<PublicPetEdge>;
  pageInfo: PageInfo;
};

export type PublicPetEdge = {
  __typename?: 'PublicPetEdge';
  cursor: Scalars['String']['output'];
  node: PublicPet;
};

export type Query = {
  __typename?: 'Query';
  availablePets: PublicPetConnection;
  soldPets: PetConnection;
  unsoldPets: PetConnection;
};


export type QueryAvailablePetsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  species?: InputMaybe<Species>;
  storeId: Scalars['ID']['input'];
};


export type QuerySoldPetsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  from: Scalars['Time']['input'];
  to: Scalars['Time']['input'];
};


export type QueryUnsoldPetsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
};

export enum Species {
  Cat = 'CAT',
  Dog = 'DOG',
  Frog = 'FROG'
}

export type AvailablePetsQueryVariables = Exact<{
  storeId: Scalars['ID']['input'];
  species?: InputMaybe<Species>;
  first?: InputMaybe<Scalars['Int']['input']>;
  after?: InputMaybe<Scalars['String']['input']>;
}>;


export type AvailablePetsQuery = { __typename?: 'Query', availablePets: { __typename?: 'PublicPetConnection', edges: Array<{ __typename?: 'PublicPetEdge', cursor: string, node: { __typename?: 'PublicPet', id: string, name: string, species: Species, ageYears: number, description: string, pictureUrl: string, status: PetStatus, createdAt: string } }>, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, endCursor?: string | null } } };

export type CheckoutMutationVariables = Exact<{
  petIds: Array<Scalars['ID']['input']> | Scalars['ID']['input'];
}>;


export type CheckoutMutation = { __typename?: 'Mutation', checkout: Array<{ __typename?: 'PublicPet', id: string, status: PetStatus }> };

export type PurchasePetMutationVariables = Exact<{
  petId: Scalars['ID']['input'];
}>;


export type PurchasePetMutation = { __typename?: 'Mutation', purchasePet: { __typename?: 'PublicPet', id: string, status: PetStatus } };


export const AvailablePetsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"AvailablePets"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"storeId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"species"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Species"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"first"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"after"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"availablePets"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"storeId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"storeId"}}},{"kind":"Argument","name":{"kind":"Name","value":"species"},"value":{"kind":"Variable","name":{"kind":"Name","value":"species"}}},{"kind":"Argument","name":{"kind":"Name","value":"first"},"value":{"kind":"Variable","name":{"kind":"Name","value":"first"}}},{"kind":"Argument","name":{"kind":"Name","value":"after"},"value":{"kind":"Variable","name":{"kind":"Name","value":"after"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"edges"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"node"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"species"}},{"kind":"Field","name":{"kind":"Name","value":"ageYears"}},{"kind":"Field","name":{"kind":"Name","value":"description"}},{"kind":"Field","name":{"kind":"Name","value":"pictureUrl"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"cursor"}}]}},{"kind":"Field","name":{"kind":"Name","value":"pageInfo"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"hasNextPage"}},{"kind":"Field","name":{"kind":"Name","value":"endCursor"}}]}}]}}]}}]} as unknown as DocumentNode<AvailablePetsQuery, AvailablePetsQueryVariables>;
export const CheckoutDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"Checkout"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"petIds"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"checkout"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"petIds"},"value":{"kind":"Variable","name":{"kind":"Name","value":"petIds"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}}]}}]}}]} as unknown as DocumentNode<CheckoutMutation, CheckoutMutationVariables>;
export const PurchasePetDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"PurchasePet"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"petId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"purchasePet"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"petId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"petId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}}]}}]}}]} as unknown as DocumentNode<PurchasePetMutation, PurchasePetMutationVariables>;