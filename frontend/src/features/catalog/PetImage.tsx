import { useState } from 'react';
import { cn } from '@/lib/cn';
import { Paw } from '@/components/brand/Paw';

interface PetImageProps {
  src: string;
  alt: string;
  /** Tailwind background classes for the placeholder/loading tint (e.g. a species tile). */
  tint?: string;
  /** Icon color class for the placeholder paw. */
  iconClassName?: string;
  className?: string;
}

/**
 * Pet photo from the API's same-origin /pictures path. Aspect-boxed (no layout shift),
 * lazy-loaded; on load failure (e.g. 404 expired key) shows a friendly paw placeholder, never a
 * broken-image glyph. The pet name stays announced via the placeholder's aria-label.
 */
export function PetImage({ src, alt, tint, iconClassName, className }: PetImageProps) {
  const [failed, setFailed] = useState(false);

  return (
    <div className={cn('relative aspect-[4/3] overflow-hidden', tint ?? 'bg-primary/10', className)}>
      {failed ? (
        <div role="img" aria-label={alt} className="flex h-full w-full items-center justify-center">
          <Paw className={cn('h-14 w-14', iconClassName ?? 'text-primary/40')} />
        </div>
      ) : (
        <img
          src={src}
          alt={alt}
          width={400}
          height={300}
          loading="lazy"
          decoding="async"
          onError={() => setFailed(true)}
          className="h-full w-full object-cover"
        />
      )}
    </div>
  );
}
