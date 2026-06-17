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
      className="flex flex-col items-center gap-4 rounded border-l-4 border-danger bg-card p-8 text-center"
    >
      <p className="font-display text-lg font-semibold">We couldn’t load the catalog</p>
      <p className="max-w-sm text-sm text-muted">{message}</p>
      <Button variant="secondary" onClick={onRetry}>
        Try again
      </Button>
    </div>
  );
}
