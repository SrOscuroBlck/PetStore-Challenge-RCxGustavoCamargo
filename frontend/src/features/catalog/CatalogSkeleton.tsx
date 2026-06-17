import { Skeleton } from '@/components/ui/Skeleton';

/** First-load placeholder grid mirroring the pet-card layout (no blank screen). */
export function CatalogSkeleton({ count = 8 }: { count?: number }) {
  return (
    <ul
      data-testid="catalog-skeleton"
      className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3"
      aria-hidden="true"
    >
      {Array.from({ length: count }).map((_, i) => (
        <li key={i} className="overflow-hidden rounded border border-border bg-card">
          <Skeleton className="aspect-[4/3] w-full rounded-none" />
          <div className="space-y-2 p-4">
            <Skeleton className="h-5 w-2/3" />
            <Skeleton className="h-4 w-1/4" />
            <Skeleton className="h-4 w-full" />
          </div>
        </li>
      ))}
    </ul>
  );
}
