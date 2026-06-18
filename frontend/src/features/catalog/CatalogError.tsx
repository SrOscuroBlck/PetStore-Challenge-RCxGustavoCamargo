import { Button } from '@/components/ui/Button';

interface CatalogErrorProps {
  message: string;
  onRetry: () => void;
}

/** Catalog load failure — shows the human-readable message and a retry. */
export function CatalogError({ message, onRetry }: CatalogErrorProps) {
  return (
    <div
      role="alert"
      className="flex flex-col items-center gap-4 rounded-2xl border-2 border-danger/40 bg-card p-10 text-center shadow-soft"
    >
      <p className="font-display text-xl font-semibold">We couldn’t load the catalog</p>
      <p className="max-w-sm text-sm text-muted">{message}</p>
      <Button variant="secondary" onClick={onRetry}>
        Try again
      </Button>
    </div>
  );
}
