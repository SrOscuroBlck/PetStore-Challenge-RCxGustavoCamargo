/** Friendly paw-print mark — the storefront's recurring motif (replaces the angular RC glyph).
    Decorative by default; pass a `title` to make it a labelled image. */
export function Paw({ className, title }: { className?: string; title?: string }) {
  return (
    <svg
      viewBox="0 0 64 64"
      className={className}
      role={title ? 'img' : undefined}
      aria-label={title}
      aria-hidden={title ? undefined : true}
    >
      <ellipse cx="20" cy="22" rx="7" ry="9" fill="currentColor" />
      <ellipse cx="44" cy="22" rx="7" ry="9" fill="currentColor" />
      <ellipse cx="10" cy="38" rx="6.5" ry="8" fill="currentColor" />
      <ellipse cx="54" cy="38" rx="6.5" ry="8" fill="currentColor" />
      <path
        fill="currentColor"
        d="M32 34c8.5 0 15 6.2 15 13.4 0 5.2-4 8.6-9.4 8.6-2.6 0-4-1.1-5.6-1.1s-3 1.1-5.6 1.1C15 56 11 52.6 11 47.4 11 40.2 23.5 34 32 34Z"
      />
    </svg>
  );
}
