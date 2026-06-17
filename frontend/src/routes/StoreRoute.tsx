import { useParams } from 'react-router-dom';
import { CatalogSection } from '@/features/catalog/CatalogSection';

/** The store page: hero + the live, paginated, filterable catalog (issue #2). */
export function StoreRoute() {
  const { storeId } = useParams();
  if (!storeId) return null; // route always provides it (/store/:storeId)

  return (
    <>
      {/* Hero — warm orange wash, big orange headline + CTA, RC mark watermark. */}
      <section className="relative overflow-hidden bg-gradient-to-br from-primary/15 via-bg to-bg">
        <svg
          viewBox="0 0 24 24"
          aria-hidden="true"
          className="pointer-events-none absolute -right-8 top-1/2 hidden h-72 w-72 -translate-y-1/2 text-primary/10 sm:block"
        >
          <path fill="currentColor" d="M2 2h7l-3.5 10L9 22H2l3.5-10z" />
          <path fill="currentColor" d="M13 2h7l-3.5 10L20 22h-7l3.5-10z" />
        </svg>
        <div className="relative mx-auto w-full max-w-6xl px-4 py-14 sm:px-6 sm:py-20">
          <p className="eyebrow">Find your companion</p>
          <h1 className="mt-3 max-w-2xl text-4xl font-bold text-primary sm:text-5xl">
            Meet the pets ready for a new home
          </h1>
          <p className="mt-5 max-w-xl text-lg text-muted">
            Browse the pets available at this store and take one home — instantly, no checkout queue.
          </p>
          <a
            href="#catalog-heading"
            className="mt-8 inline-flex items-center rounded-sm bg-primary px-6 py-3 font-display text-sm font-semibold uppercase tracking-wide text-primary-fg transition-colors hover:bg-primary-strong"
          >
            Browse pets
          </a>
        </div>
      </section>

      <CatalogSection storeId={storeId} />
    </>
  );
}
