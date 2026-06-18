/** Friendly filled animal-face icons (currentColor), used for species badges + the filter.
    Decorative — callers supply the accessible label alongside. */

interface IconProps {
  className?: string;
}

export function CatIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" className={className} fill="currentColor" aria-hidden="true">
      <path d="M2 2.5l5 3.6A12.6 12.6 0 0 1 12 5c1.8 0 3.5.4 5 1.1l5-3.6-1.3 7.3c.8 1.1 1.3 2.4 1.3 3.7 0 4.4-4.5 7.5-10 7.5S2 17.9 2 13.5c0-1.3.5-2.6 1.3-3.7L2 2.5zm6.5 11a1.2 1.2 0 1 0 0-2.4 1.2 1.2 0 0 0 0 2.4zm7 0a1.2 1.2 0 1 0 0-2.4 1.2 1.2 0 0 0 0 2.4zM12 15c-.9 0-1.6.5-1.6 1s.7 1 1.6 1 1.6-.5 1.6-1-.7-1-1.6-1z" />
    </svg>
  );
}

export function DogIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" className={className} fill="currentColor" aria-hidden="true">
      <path d="M5 3c-1.7 0-3 1.7-3 4.6 0 1.9.6 3.4 1.6 4.3-.4.8-.6 1.7-.6 2.6C3 18.6 7 22 12 22s9-3.4 9-7.5c0-.9-.2-1.8-.6-2.6 1-.9 1.6-2.4 1.6-4.3C22 4.7 20.7 3 19 3c-1.6 0-2.9 1.4-3.4 3.5A6.6 6.6 0 0 0 12 5.4c-1.3 0-2.5.4-3.6 1.1C7.9 4.4 6.6 3 5 3zm3.5 9.5a1.2 1.2 0 1 0 0-2.4 1.2 1.2 0 0 0 0 2.4zm7 0a1.2 1.2 0 1 0 0-2.4 1.2 1.2 0 0 0 0 2.4zM12 14.5c-1.2 0-2.2.7-2.2 1.6 0 .9 1 1.4 2.2 1.4s2.2-.5 2.2-1.4c0-.9-1-1.6-2.2-1.6z" />
    </svg>
  );
}

export function FrogIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" className={className} fill="currentColor" aria-hidden="true">
      <path d="M6.5 3A3.5 3.5 0 0 0 3 6.5c0 .8.3 1.6.8 2.2A8.6 8.6 0 0 0 2 13.5C2 17.6 6.5 21 12 21s10-3.4 10-7.5c0-1.7-.7-3.3-1.8-4.6.5-.6.8-1.4.8-2.4A3.5 3.5 0 0 0 17.5 3c-1.5 0-2.8 1-3.3 2.4A9.7 9.7 0 0 0 12 5.2c-.8 0-1.5.1-2.2.2A3.5 3.5 0 0 0 6.5 3zM7 8.2a1.4 1.4 0 1 0 0-2.8 1.4 1.4 0 0 0 0 2.8zm10 0a1.4 1.4 0 1 0 0-2.8 1.4 1.4 0 0 0 0 2.8zM8 15h8a4 4 0 0 1-8 0z" />
    </svg>
  );
}
