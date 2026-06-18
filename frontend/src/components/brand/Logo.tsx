import { cn } from '@/lib/cn';
import { Paw } from './Paw';

/** Wordmark: a friendly orange paw badge + a soft rounded two-tone display wordmark. */
export function Logo({ className }: { className?: string }) {
  return (
    <span className={cn('inline-flex items-center gap-2.5', className)}>
      <span className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-primary">
        <Paw className="h-5 w-5 text-primary-fg" />
      </span>
      <span className="font-display text-xl font-semibold lowercase tracking-tight">
        pet<span className="text-primary">store</span>
      </span>
    </span>
  );
}
