import { useState } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { Button } from '@/components/ui/Button';
import { useCart } from './useCart';
import { useAuth } from '@/features/auth/useAuth';
import { useCheckout } from '@/features/purchase/useCheckout';
import { CartLineItem } from './CartLineItem';

interface CartDrawerProps {
  /** Catalog refetch — used to reconcile after a failed checkout. */
  refetch: () => void;
}

/** Right-side cart drawer (Radix Dialog: focus trap + Esc + labelling for free). */
export function CartDrawer({ refetch }: CartDrawerProps) {
  const { items, count, isOpen, close, remove } = useCart();
  const { ensureSignedIn } = useAuth();
  const { checkout, loading } = useCheckout(refetch);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const handleCheckout = async () => {
    // Placing the order requires sign-in (ADR-0006); browsing/cart stay open.
    if (!(await ensureSignedIn())) return;
    setErrorMessage(null);
    const result = await checkout();
    // On UNAVAILABLE the message names every unavailable pet — surface it verbatim.
    if (!result.ok && result.errorMessage) setErrorMessage(result.errorMessage);
  };

  return (
    <Dialog.Root
      open={isOpen}
      onOpenChange={(open) => {
        if (!open) {
          setErrorMessage(null);
          close();
        }
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-black/50" />
        <Dialog.Content className="fixed inset-y-0 right-0 z-50 flex w-full max-w-md flex-col bg-bg shadow-xl focus:outline-none">
          <div className="flex items-center justify-between border-b border-border bg-ink px-5 py-4 text-ink-fg">
            <Dialog.Title className="font-display text-lg font-bold">Your cart</Dialog.Title>
            <Dialog.Close
              aria-label="Close cart"
              className="rounded-sm p-1 transition-colors hover:bg-white/10"
            >
              <svg
                viewBox="0 0 24 24"
                aria-hidden="true"
                className="h-6 w-6"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
              >
                <path d="M18 6 6 18M6 6l12 12" />
              </svg>
            </Dialog.Close>
          </div>
          <Dialog.Description className="sr-only">
            Review the pets in your cart and check out.
          </Dialog.Description>

          {count === 0 ? (
            <div className="flex flex-1 flex-col items-center justify-center gap-2 p-8 text-center">
              <p className="font-display text-lg font-semibold">Your cart is empty</p>
              <p className="text-sm text-muted">Add a pet from the catalog to get started.</p>
            </div>
          ) : (
            <ul className="flex-1 divide-y divide-border overflow-y-auto px-5">
              {items.map((item) => (
                <CartLineItem key={item.id} item={item} onRemove={remove} />
              ))}
            </ul>
          )}

          {errorMessage ? (
            <p role="alert" className="border-t border-danger/40 bg-danger/5 px-5 py-3 text-sm text-danger">
              {errorMessage}
            </p>
          ) : null}

          <div className="border-t border-border p-5">
            <Button
              variant="primary"
              className="w-full"
              disabled={count === 0 || loading}
              onClick={() => void handleCheckout()}
            >
              {loading ? 'Checking out…' : count > 0 ? `Checkout (${count})` : 'Checkout'}
            </Button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
