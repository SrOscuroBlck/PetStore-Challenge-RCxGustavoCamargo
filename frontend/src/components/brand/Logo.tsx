import { cn } from '@/lib/cn';

/** Wordmark in the Robotic Crew idiom: angular orange mark + two-tone display wordmark. */
export function Logo({ className }: { className?: string }) {
  return (
    <span className={cn('inline-flex items-center gap-2', className)}>
      <svg viewBox="0 0 24 24" className="h-7 w-7 text-primary" aria-hidden="true">
        <path fill="currentColor" d="M2 2h7l-3.5 10L9 22H2l3.5-10z" />
        <path fill="currentColor" opacity="0.65" d="M13 2h7l-3.5 10L20 22h-7l3.5-10z" />
      </svg>
      <span className="font-display text-lg font-bold uppercase tracking-wide">
        Pet<span className="text-primary">store</span>
      </span>
    </span>
  );
}
