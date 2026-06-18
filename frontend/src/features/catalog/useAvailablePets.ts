import { useCallback } from 'react';
import { NetworkStatus, useQuery } from '@apollo/client';
import { AvailablePetsDocument } from '@/__generated__/graphql';
import type { AvailablePetsQuery, Species } from '@/__generated__/graphql';
import { mapApolloError } from '@/api/errors/mapError';
import type { MappedError } from '@/api/errors/mapError';

/** Catalog page size — must stay ≤ 100 (server complexity cap). */
export const PAGE_SIZE = 12;

/** One pet node, derived from the generated query type (no hand-rolled shape). */
export type CatalogPet = AvailablePetsQuery['availablePets']['edges'][number]['node'];

export interface UseAvailablePetsResult {
  pets: CatalogPet[];
  /** True only on the first-ever load (drives the skeleton, not background refetches). */
  loading: boolean;
  loadingMore: boolean;
  error: MappedError | undefined;
  hasNextPage: boolean;
  loadMore: () => void;
  refetch: () => void;
}

/**
 * Browse a store's available pets. `species` null = all (the backend treats null/absent as
 * unfiltered). Pages merge via the relay cache keyed on [storeId, species], so switching
 * filters keeps each set's pagination independent (null is its own "All" bucket).
 */
export function useAvailablePets(
  storeId: string,
  species: Species | null,
): UseAvailablePetsResult {
  const { data, error, networkStatus, fetchMore, refetch } = useQuery(AvailablePetsDocument, {
    variables: { storeId, species, first: PAGE_SIZE },
    notifyOnNetworkStatusChange: true,
  });

  const connection = data?.availablePets;
  const pets = connection?.edges.map((edge) => edge.node) ?? [];
  const hasNextPage = connection?.pageInfo.hasNextPage ?? false;
  const endCursor = connection?.pageInfo.endCursor ?? null;

  const loadMore = useCallback(() => {
    if (!hasNextPage || networkStatus === NetworkStatus.fetchMore || !endCursor) return;
    void fetchMore({ variables: { after: endCursor } });
  }, [hasNextPage, networkStatus, endCursor, fetchMore]);

  const doRefetch = useCallback(() => {
    void refetch();
  }, [refetch]);

  return {
    pets,
    loading: networkStatus === NetworkStatus.loading,
    loadingMore: networkStatus === NetworkStatus.fetchMore,
    error: error ? mapApolloError(error) : undefined,
    hasNextPage,
    loadMore,
    refetch: doRefetch,
  };
}
