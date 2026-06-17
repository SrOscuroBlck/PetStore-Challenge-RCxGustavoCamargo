import { speciesLabel } from '@/features/catalog/speciesLabel';
import type { CartItem } from './cartContext';

interface CartLineItemProps {
  item: CartItem;
  onRemove: (id: string) => void;
}

export function CartLineItem({ item, onRemove }: CartLineItemProps) {
  return (
    <li className="flex items-center gap-3 py-3">
      {/* Decorative thumbnail — the name is the accessible label, alongside. */}
      <img
        src={item.pictureUrl}
        alt=""
        width={56}
        height={56}
        loading="lazy"
        className="h-14 w-14 shrink-0 rounded bg-primary/10 object-cover"
      />
      <div className="min-w-0 flex-1">
        <p className="truncate font-display font-semibold">{item.name}</p>
        <p className="text-sm text-muted">{speciesLabel(item.species)}</p>
      </div>
      <button
        type="button"
        onClick={() => onRemove(item.id)}
        aria-label={`Remove ${item.name}`}
        className="rounded-sm p-2 text-muted transition-colors hover:text-danger"
      >
        <svg
          viewBox="0 0 24 24"
          aria-hidden="true"
          className="h-5 w-5"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M3 6h18M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2m2 0v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6" />
        </svg>
      </button>
    </li>
  );
}
