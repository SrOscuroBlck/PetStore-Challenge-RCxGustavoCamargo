import { useParams } from 'react-router-dom';
import { CatalogSection } from '@/features/catalog/CatalogSection';
import { Paw } from '@/components/brand/Paw';

/** The store page: friendly hero + the live, paginated, filterable catalog (issue #2). */
export function StoreRoute() {
  const { storeId } = useParams();
  if (!storeId) return null; // route always provides it (/store/:storeId)

  return (
    <>
      {/* Hero — warm cream-to-peach wash with floating paw prints + a big friendly headline. */}
      <section className="relative overflow-hidden bg-gradient-to-br from-primary/20 via-primary/5 to-bg">
        <Paw className="pointer-events-none absolute -right-6 top-10 hidden h-40 w-40 rotate-12 text-primary/10 sm:block" />
        <Paw className="pointer-events-none absolute right-44 top-44 hidden h-20 w-20 -rotate-12 text-primary/10 lg:block" />
        <Paw className="pointer-events-none absolute -left-6 -bottom-6 h-28 w-28 rotate-6 text-primary/10" />
        <div className="relative mx-auto w-full max-w-6xl px-4 py-16 sm:px-6 sm:py-24">
          <p className="eyebrow flex items-center gap-2">
            <Paw className="h-4 w-4 text-primary" /> Find your new best friend
          </p>
          <h1 className="mt-4 max-w-2xl text-4xl font-bold leading-[1.05] text-fg sm:text-6xl">
            Meet the pets ready for a <span className="text-primary">new home</span>
          </h1>
          <p className="mt-5 max-w-xl text-lg text-muted">
            Browse the happy faces available at this store and take one home — instantly, no checkout
            queue.
          </p>
          <a
            href="#catalog-heading"
            className="mt-8 inline-flex items-center gap-2 rounded-full bg-primary px-7 py-3.5 font-display text-sm font-semibold tracking-wide text-primary-fg shadow-soft transition-all hover:bg-primary-strong hover:shadow-lift motion-safe:active:scale-95"
          >
            <Paw className="h-4 w-4" /> Browse pets
          </a>
        </div>
      </section>

      <CatalogSection storeId={storeId} />
    </>
  );
}
