import { describe, expect, it } from 'vitest';
import { AvailablePetsDocument, PetStatus, Species } from '@/__generated__/graphql';
import type { AvailablePetsQuery } from '@/__generated__/graphql';
import { makeCache } from '@/api/apollo/cache';
import { removePetsFromCatalog } from '../catalogCache';

type Node = AvailablePetsQuery['availablePets']['edges'][number]['node'];
type Connection = AvailablePetsQuery['availablePets'];

const STORE = 's1';

function pet(id: string, species: Species): Node {
  return {
    __typename: 'PublicPet',
    id,
    name: `Pet ${id}`,
    species,
    ageYears: 1,
    description: '',
    pictureUrl: `/p/${id}`,
    status: PetStatus.Available,
    createdAt: '2026-01-01T00:00:00Z',
  };
}

function connection(nodes: Node[]): Connection {
  return {
    __typename: 'PublicPetConnection',
    edges: nodes.map((node) => ({ __typename: 'PublicPetEdge', cursor: `c${node.id}`, node })),
    pageInfo: { __typename: 'PageInfo', hasNextPage: false, endCursor: null },
  };
}

function idsIn(cache: ReturnType<typeof makeCache>, species: Species | null): string[] {
  const data = cache.readQuery({
    query: AvailablePetsDocument,
    variables: { storeId: STORE, species, first: 12 },
  });
  return data?.availablePets.edges.map((edge) => edge.node.id) ?? [];
}

describe('removePetsFromCatalog', () => {
  it('removes a pet from every [storeId, species] bucket at once', () => {
    const cache = makeCache();
    // Pet "1" lives in both the "All" bucket and the CAT bucket.
    cache.writeQuery({
      query: AvailablePetsDocument,
      variables: { storeId: STORE, species: null, first: 12 },
      data: { availablePets: connection([pet('1', Species.Cat), pet('2', Species.Dog)]) },
    });
    cache.writeQuery({
      query: AvailablePetsDocument,
      variables: { storeId: STORE, species: Species.Cat, first: 12 },
      data: { availablePets: connection([pet('1', Species.Cat)]) },
    });

    removePetsFromCatalog(cache, ['1']);

    expect(idsIn(cache, null)).toEqual(['2']); // pet 1 gone, pet 2 survives
    expect(idsIn(cache, Species.Cat)).toEqual([]); // pet 1 gone from the CAT bucket too
  });
});
