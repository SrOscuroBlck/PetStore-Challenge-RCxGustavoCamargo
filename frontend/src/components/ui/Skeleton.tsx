import { cn } from '@/lib/cn';

/** Decorative loading placeholder. Hidden from assistive tech. */
export function Skeleton({ className }: { className?: string }) {
  return <div aria-hidden className={cn('animate-pulse rounded bg-border/60', className)} />;
}
