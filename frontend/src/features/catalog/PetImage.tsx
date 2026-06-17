import { useState } from 'react';
import { cn } from '@/lib/cn';

interface PetImageProps {
  src: string;
  alt: string;
  className?: string;
}

/**
 * Pet photo from the API's same-origin /pictures path. Aspect-boxed (no layout shift),
 * lazy-loaded; on load failure (e.g. 404 expired key) shows a branded placeholder, never a
 * broken-image glyph. The pet name stays announced via the placeholder's aria-label.
 */
export function PetImage({ src, alt, className }: PetImageProps) {
  const [failed, setFailed] = useState(false);

  return (
    <div className={cn('relative aspect-[4/3] overflow-hidden bg-primary/10', className)}>
      {failed ? (
        <div role="img" aria-label={alt} className="flex h-full w-full items-center justify-center">
          <svg viewBox="0 0 24 24" aria-hidden="true" className="h-12 w-12 text-primary/40">
            <path fill="currentColor" d="M2 2h7l-3.5 10L9 22H2l3.5-10z" />
            <path fill="currentColor" opacity="0.65" d="M13 2h7l-3.5 10L20 22h-7l3.5-10z" />
          </svg>
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
