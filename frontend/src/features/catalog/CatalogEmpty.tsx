import { Paw } from '@/components/brand/Paw';

/** Friendly empty state — covers a store with no available pets, an unknown store, and a
    species filter with no matches (all yield an empty connection, never an error). */
export function CatalogEmpty() {
  return (
    <div className="flex flex-col items-center gap-4 rounded-2xl border-2 border-dashed border-border bg-card/70 p-14 text-center">
      <span className="inline-flex h-16 w-16 items-center justify-center rounded-full bg-primary/10">
        <Paw className="h-9 w-9 text-primary/60" />
      </span>
      <p className="font-display text-xl font-semibold">No pets available right now</p>
      <p className="max-w-sm text-sm text-muted">
        Check back soon — new companions are listed all the time.
      </p>
    </div>
  );
}
