import { useCart } from './useCart';

/** Header cart trigger with an orange count badge. Opens the drawer. */
export function CartButton() {
  const { count, open } = useCart();

  return (
    <button
      type="button"
      onClick={open}
      aria-label={`Cart, ${count} ${count === 1 ? 'item' : 'items'}`}
      className="relative inline-flex h-10 w-10 items-center justify-center rounded-sm text-ink-fg transition-colors hover:bg-white/10"
    >
      <svg
        viewBox="0 0 24 24"
        aria-hidden="true"
        className="h-6 w-6"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <circle cx="9" cy="21" r="1" />
        <circle cx="20" cy="21" r="1" />
        <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6" />
      </svg>
      {count > 0 ? (
        <span
          aria-hidden="true"
          className="absolute -right-1 -top-1 inline-flex h-5 min-w-[1.25rem] items-center justify-center rounded-full bg-primary px-1 text-xs font-bold text-primary-fg"
        >
          {count}
        </span>
      ) : null}
    </button>
  );
}
