import { motion } from 'framer-motion';
import { PetStatus } from '@/__generated__/graphql';
import { cn } from '@/lib/cn';
import type { CatalogPet } from './useAvailablePets';
import { formatAge } from './formatAge';
import { speciesLabel } from './speciesLabel';
import { PetImage } from './PetImage';
import { PetCardActions } from './PetCardActions';

interface PetCardProps {
  pet: CatalogPet;
  refetch: () => void;
}

const itemVariants = {
  hidden: { opacity: 0, y: 8 },
  visible: { opacity: 1, y: 0 },
};

export function PetCard({ pet, refetch }: PetCardProps) {
  // Defensive: availablePets only ever returns AVAILABLE, but keep the component correct
  // if reused with mixed data.
  const sold = pet.status !== PetStatus.Available;

  return (
    <motion.li
      variants={itemVariants}
      className={cn(
        'flex flex-col overflow-hidden rounded border border-border bg-card transition-[transform,border-color]',
        sold ? 'opacity-60' : 'hover:border-primary motion-safe:hover:-translate-y-1',
      )}
    >
      <div className="relative">
        <PetImage src={pet.pictureUrl} alt={pet.name} className="clip-corner-tr" />
        <span className="absolute left-3 top-3 rounded-sm bg-primary px-2 py-0.5 font-display text-xs font-semibold uppercase tracking-wide text-primary-fg">
          {speciesLabel(pet.species)}
        </span>
        {sold ? (
          <span className="absolute right-3 top-3 rounded-sm bg-ink/80 px-2 py-0.5 font-display text-xs font-semibold uppercase tracking-wide text-ink-fg">
            Sold
          </span>
        ) : null}
      </div>
      <div className="flex flex-1 flex-col p-4">
        <h3 className="font-display text-lg font-semibold">{pet.name}</h3>
        <p className="mt-0.5 text-sm text-muted">{formatAge(pet.ageYears)}</p>
        <p className="mt-2 line-clamp-2 text-sm text-muted">{pet.description}</p>
        <div data-testid="pet-card-actions" className="mt-4">
          {sold ? null : <PetCardActions pet={pet} refetch={refetch} />}
        </div>
      </div>
    </motion.li>
  );
}
