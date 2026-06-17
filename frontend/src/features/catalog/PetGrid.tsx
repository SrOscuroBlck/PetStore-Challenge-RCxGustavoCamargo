import { motion } from 'framer-motion';
import { useReducedMotion } from '@/lib/a11y/useReducedMotion';
import { Button } from '@/components/ui/Button';
import type { CatalogPet } from './useAvailablePets';
import { PetCard } from './PetCard';
import { useInfiniteScroll } from './useInfiniteScroll';

interface PetGridProps {
  pets: CatalogPet[];
  hasNextPage: boolean;
  loadingMore: boolean;
  onLoadMore: () => void;
  refetch: () => void;
}

export function PetGrid({ pets, hasNextPage, loadingMore, onLoadMore, refetch }: PetGridProps) {
  const reduce = useReducedMotion();
  const sentinelRef = useInfiniteScroll<HTMLDivElement>({
    onLoadMore,
    enabled: hasNextPage && !loadingMore,
  });

  return (
    <div>
      <motion.ul
        className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3"
        initial={reduce ? false : 'hidden'}
        animate="visible"
        variants={{ visible: { transition: { staggerChildren: reduce ? 0 : 0.05 } } }}
      >
        {pets.map((pet) => (
          <PetCard key={pet.id} pet={pet} refetch={refetch} />
        ))}
      </motion.ul>

      <div aria-live="polite" className="sr-only">
        {loadingMore ? 'Loading more pets…' : !hasNextPage ? 'All pets loaded.' : ''}
      </div>

      {hasNextPage ? (
        <div className="mt-10 flex flex-col items-center gap-2">
          {/* Sentinel: auto-loads when near the viewport; the button is the keyboard/no-observer fallback. */}
          <div ref={sentinelRef} aria-hidden="true" />
          <Button variant="secondary" onClick={onLoadMore} disabled={loadingMore}>
            {loadingMore ? 'Loading…' : 'Load more'}
          </Button>
        </div>
      ) : null}
    </div>
  );
}
