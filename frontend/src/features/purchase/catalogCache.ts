import type { ApolloCache, Reference } from '@apollo/client';

type EdgeRef = { node: Reference; cursor?: string };
type ConnectionValue = { edges?: readonly EdgeRef[]; [field: string]: unknown };

/**
 * Remove the given pet ids from EVERY `availablePets` relay bucket. The field is keyed on
 * [storeId, species], so a pet may live in the "All" bucket and its species bucket at once;
 * `cache.modify` runs the modifier once per stored field key, filtering edges in all of them.
 * Used by both single-purchase and checkout for instant, optimistic removal.
 */
export function removePetsFromCatalog(cache: ApolloCache<unknown>, petIds: readonly string[]): void {
  const ids = new Set(petIds.map(String));
  cache.modify({
    fields: {
      availablePets(existing, { readField }) {
        const connection = existing as ConnectionValue | undefined;
        if (!connection?.edges) return existing;
        const edges = connection.edges.filter((edge) => !ids.has(String(readField('id', edge.node))));
        if (edges.length === connection.edges.length) return existing;
        return { ...connection, edges };
      },
    },
  });
}
