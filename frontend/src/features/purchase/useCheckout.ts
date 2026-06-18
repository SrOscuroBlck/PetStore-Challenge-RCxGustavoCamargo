import { useCallback, useState } from 'react';
import { useMutation } from '@apollo/client';
import { CheckoutDocument, PetStatus } from '@/__generated__/graphql';
import { mapApolloError } from '@/api/errors/mapError';
import { useToast } from '@/components/ui/Toast/useToast';
import { useCart } from '@/features/cart/useCart';
import { removePetsFromCatalog } from './catalogCache';

export interface CheckoutResult {
  ok: boolean;
  /** On failure, the server's human-readable message (for UNAVAILABLE it names every unavailable pet). */
  errorMessage?: string;
}

/**
 * Atomic cart checkout. Optimistically removes all carted pets; on success clears the cart.
 * On UNAVAILABLE the server message names every unavailable pet — we surface it verbatim and
 * refetch to reconcile the catalog to true availability.
 */
export function useCheckout(refetch: () => void) {
  const [mutate] = useMutation(CheckoutDocument);
  const { show } = useToast();
  const { items, clear } = useCart();
  const [loading, setLoading] = useState(false);

  const checkout = useCallback(async (): Promise<CheckoutResult> => {
    if (items.length === 0) return { ok: false };
    const petIds = items.map((item) => item.id);
    setLoading(true);
    try {
      await mutate({
        variables: { petIds },
        optimisticResponse: {
          checkout: petIds.map((id) => ({ __typename: 'PublicPet' as const, id, status: PetStatus.Sold })),
        },
        update(cache) {
          removePetsFromCatalog(cache, petIds);
        },
      });
      clear();
      const noun = petIds.length === 1 ? 'pet' : 'pets';
      show({ description: `Adopted ${petIds.length} ${noun}!`, variant: 'success' });
      return { ok: true };
    } catch (error) {
      const mapped = mapApolloError(error);
      show({ title: 'Checkout failed', description: mapped.userMessage, variant: 'error' });
      if (mapped.code === 'UNAVAILABLE') refetch();
      return { ok: false, errorMessage: mapped.userMessage };
    } finally {
      setLoading(false);
    }
  }, [items, mutate, show, clear, refetch]);

  return { checkout, loading };
}
