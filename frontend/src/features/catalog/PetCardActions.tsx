import { Button } from '@/components/ui/Button';
import { useCart } from '@/features/cart/useCart';
import { useAuth } from '@/features/auth/useAuth';
import { usePurchasePet } from '@/features/purchase/usePurchasePet';
import type { CatalogPet } from './useAvailablePets';

interface PetCardActionsProps {
  pet: CatalogPet;
  refetch: () => void;
}

/** Buy (instant) + Add-to-cart. When already carted, the secondary button removes it (no duplicates). */
export function PetCardActions({ pet, refetch }: PetCardActionsProps) {
  const { has, add, remove } = useCart();
  const { ensureSignedIn } = useAuth();
  const { purchase, pendingId } = usePurchasePet(refetch);
  const inCart = has(pet.id);
  const buying = pendingId === pet.id;

  // Browsing is open; placing an order requires sign-in (ADR-0006).
  const handleBuy = async () => {
    if (await ensureSignedIn()) await purchase(pet);
  };

  return (
    <div className="flex items-center gap-2">
      <Button variant="primary" className="flex-1" disabled={buying} onClick={() => void handleBuy()}>
        {buying ? 'Buying…' : 'Buy'}
      </Button>
      {inCart ? (
        <Button
          variant="secondary"
          className="flex-1"
          onClick={() => remove(pet.id)}
          aria-label={`Remove ${pet.name} from cart`}
        >
          In cart ✓
        </Button>
      ) : (
        <Button
          variant="secondary"
          className="flex-1"
          onClick={() =>
            add({ id: pet.id, name: pet.name, species: pet.species, pictureUrl: pet.pictureUrl })
          }
        >
          Add to cart
        </Button>
      )}
    </div>
  );
}
