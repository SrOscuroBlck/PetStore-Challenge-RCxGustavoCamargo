import { Logo } from '@/components/brand/Logo';

/** Friendly empty state — covers a store with no available pets, an unknown store, and a
    species filter with no matches (all yield an empty connection, never an error). */
export function CatalogEmpty() {
  return (
    <div className="clip-corner flex flex-col items-center gap-3 border border-dashed border-border bg-card/60 p-12 text-center">
      <Logo className="opacity-40" />
      <p className="font-display text-lg font-semibold">No pets available right now</p>
      <p className="max-w-sm text-sm text-muted">
        Check back soon — new companions are listed all the time.
      </p>
    </div>
  );
}
