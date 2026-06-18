import { InMemoryCache } from '@apollo/client';
import { relayStylePagination } from '@apollo/client/utilities';

/**
 * Relay cursor pagination is configured now (keyed per store) even though infinite scroll
 * lands in issue #2 — so the cache merges pages correctly from the first catalog query.
 */
export function makeCache(): InMemoryCache {
  return new InMemoryCache({
    typePolicies: {
      Query: {
        fields: {
          // Key by store AND species so each filter (incl. "All" = no species) paginates
          // independently and switching filters never shows stale cross-filter results.
          availablePets: relayStylePagination(['storeId', 'species']),
        },
      },
      PublicPet: { keyFields: ['id'] },
    },
  });
}
