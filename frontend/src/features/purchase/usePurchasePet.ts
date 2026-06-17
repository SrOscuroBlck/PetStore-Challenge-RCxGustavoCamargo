import { useCallback, useState } from 'react';
import { useMutation } from '@apollo/client';
import { PurchasePetDocument, PetStatus } from '@/__generated__/graphql';
import { mapApolloError } from '@/api/errors/mapError';
import { useToast } from '@/components/ui/Toast/useToast';
import { useCart } from '@/features/cart/useCart';
import type { CatalogPet } from '@/features/catalog/useAvailablePets';
import { removePetsFromCatalog } from './catalogCache';

/**
 * Instant single purchase. Optimistically removes the pet from the catalog (card animates out);
 * Apollo rolls the optimistic layer back automatically on a network error. On UNAVAILABLE the
 * pet is genuinely gone, so we reconcile via refetch instead of letting the rollback re-add it.
 */
export function usePurchasePet(refetch: () => void) {
  const [mutate] = useMutation(PurchasePetDocument);
  const { show } = useToast();
  const { remove: removeFromCart } = useCart();
  const [pendingId, setPendingId] = useState<string | null>(null);

  const purchase = useCallback(
    async (pet: CatalogPet) => {
      setPendingId(pet.id);
      try {
        await mutate({
          variables: { petId: pet.id },
          optimisticResponse: {
            purchasePet: { __typename: 'PublicPet', id: pet.id, status: PetStatus.Sold },
          },
          update(cache) {
            removePetsFromCatalog(cache, [pet.id]);
          },
        });
        removeFromCart(pet.id);
        show({ description: `${pet.name} is yours!`, variant: 'success' });
      } catch (error) {
        const mapped = mapApolloError(error);
        show({ title: 'Purchase failed', description: mapped.userMessage, variant: 'error' });
        if (mapped.code === 'UNAVAILABLE') refetch();
      } finally {
        setPendingId(null);
      }
    },
    [mutate, show, removeFromCart, refetch],
  );

  return { purchase, pendingId };
}
