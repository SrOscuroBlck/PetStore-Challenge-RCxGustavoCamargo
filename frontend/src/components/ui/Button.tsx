import { forwardRef, type ButtonHTMLAttributes } from 'react';
import { cn } from '@/lib/cn';

type Variant = 'primary' | 'secondary' | 'ghost' | 'onDark';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
}

// Uppercase display labels echo the RC CTA style. Primary uses ink-on-orange (AA ~7:1);
// orange is never used as small text on a light surface (see docs/DESIGN.md).
const variants: Record<Variant, string> = {
  primary: 'bg-primary text-primary-fg hover:bg-primary-strong',
  secondary: 'border border-border text-fg hover:border-primary hover:bg-primary/5',
  ghost: 'text-fg hover:bg-card',
  onDark: 'border border-white/30 text-ink-fg hover:border-primary hover:bg-white/5',
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
        'inline-flex items-center justify-center gap-2 rounded-sm px-5 py-2.5 font-display text-sm font-semibold uppercase tracking-wide transition-colors disabled:cursor-not-allowed disabled:opacity-60',
        variants[variant],
        className,
      )}
      {...rest}
    />
  );
});
