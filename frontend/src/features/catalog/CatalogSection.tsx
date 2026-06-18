import { useState } from 'react';
import type { Species } from '@/__generated__/graphql';
import { useAvailablePets } from './useAvailablePets';
import { SpeciesFilter } from './SpeciesFilter';
import { PetGrid } from './PetGrid';
import { CatalogSkeleton } from './CatalogSkeleton';
import { CatalogEmpty } from './CatalogEmpty';
import { CatalogError } from './CatalogError';
import { CartDrawer } from '@/features/cart/CartDrawer';
import { LoginDialog } from '@/features/auth/LoginDialog';

interface CatalogSectionProps {
  storeId: string;
}

/** The live catalog: species filter + the browse/paginate state machine. */
export function CatalogSection({ storeId }: CatalogSectionProps) {
  const [species, setSpecies] = useState<Species | null>(null);
  const { pets, loading, loadingMore, error, hasNextPage, loadMore, refetch } = useAvailablePets(
    storeId,
    species,
  );

  return (
    <section
      aria-labelledby="catalog-heading"
      className="mx-auto w-full max-w-6xl px-4 pb-20 pt-14 sm:px-6"
    >
      <p className="eyebrow">Available now</p>
      <h2 id="catalog-heading" className="rule-soft mt-2 text-3xl font-bold text-primary">
        Meet the pets
      </h2>

      <div className="mt-6">
        <SpeciesFilter value={species} onChange={setSpecies} />
      </div>

      <div className="mt-8">
        {loading ? (
          <CatalogSkeleton />
        ) : error && pets.length === 0 ? (
          <CatalogError message={error.userMessage} onRetry={refetch} />
        ) : pets.length === 0 ? (
          <CatalogEmpty />
        ) : (
          <PetGrid
            pets={pets}
            hasNextPage={hasNextPage}
            loadingMore={loadingMore}
            onLoadMore={loadMore}
            refetch={refetch}
          />
        )}
      </div>

      {/* Cart drawer + login dialog live here (inside the router) so they have the store id
          and the catalog's refetch for post-checkout reconcile. */}
      <CartDrawer refetch={refetch} />
      <LoginDialog />
    </section>
  );
}
