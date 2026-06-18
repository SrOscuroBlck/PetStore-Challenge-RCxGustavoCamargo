import { forwardRef, type ButtonHTMLAttributes } from 'react';
import { cn } from '@/lib/cn';

type Variant = 'primary' | 'secondary' | 'ghost' | 'onDark';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
}

// Friendly rounded pill CTAs. Primary uses dark-on-orange (AA); orange is never used as
// small text on a light surface (see docs/DESIGN.md).
const variants: Record<Variant, string> = {
  primary: 'bg-primary text-primary-fg shadow-soft hover:bg-primary-strong hover:shadow-lift',
  secondary: 'border-2 border-border bg-card text-fg hover:border-primary hover:bg-primary/5',
  ghost: 'text-fg hover:bg-primary/5',
  onDark: 'border-2 border-white/30 text-ink-fg hover:border-primary hover:bg-white/5',
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { variant = 'primary', className, type = 'button', ...rest },
  ref,
) {
  return (
    <button
      ref={ref}
      type={type}
      className={cn(
        'inline-flex items-center justify-center gap-2 rounded-full px-6 py-2.5 font-display text-sm font-semibold tracking-wide transition-all motion-safe:active:scale-95 disabled:cursor-not-allowed disabled:opacity-60',
        variants[variant],
        className,
      )}
      {...rest}
    />
  );
});
