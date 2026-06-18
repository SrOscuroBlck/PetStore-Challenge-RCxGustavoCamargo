import { ApolloClient, from, type NormalizedCacheObject } from '@apollo/client';
import { makeCache } from './cache';
import { authLink, httpLink } from './links';

/**
 * authLink attaches the signed-in customer's credential (ADR-0006); when signed out it sends
 * none and the gateway injects the ambient browse credential.
 */
export function makeApolloClient(): ApolloClient<NormalizedCacheObject> {
  return new ApolloClient({
    link: from([authLink, httpLink]),
    cache: makeCache(),
    defaultOptions: {
      watchQuery: { fetchPolicy: 'cache-and-network', nextFetchPolicy: 'cache-first' },
    },
  });
}
