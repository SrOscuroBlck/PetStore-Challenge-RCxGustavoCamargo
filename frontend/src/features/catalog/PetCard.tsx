import { motion } from 'framer-motion';
import { PetStatus } from '@/__generated__/graphql';
import { cn } from '@/lib/cn';
import type { CatalogPet } from './useAvailablePets';
import { formatAge } from './formatAge';
import { speciesLabel } from './speciesLabel';
import { speciesTheme } from './speciesTheme';
import { PetImage } from './PetImage';
import { PetCardActions } from './PetCardActions';

interface PetCardProps {
  pet: CatalogPet;
  refetch: () => void;
}

const itemVariants = {
  hidden: { opacity: 0, y: 10 },
  visible: { opacity: 1, y: 0 },
};

export function PetCard({ pet, refetch }: PetCardProps) {
  // Defensive: availablePets only ever returns AVAILABLE, but keep the component correct
  // if reused with mixed data.
  const sold = pet.status !== PetStatus.Available;
  const theme = speciesTheme(pet.species);

  return (
    <motion.li
      variants={itemVariants}
      className={cn(
        'group flex flex-col overflow-hidden rounded-2xl border border-border bg-card shadow-soft transition-[transform,box-shadow]',
        sold ? 'opacity-60' : 'hover:shadow-lift motion-safe:hover:-translate-y-1.5',
      )}
    >
      <div className="relative">
        <PetImage
          src={pet.pictureUrl}
          alt={pet.name}
          tint={theme.tile}
          iconClassName={theme.ink}
        />
        <span
          className={cn(
            'absolute left-3 top-3 inline-flex items-center gap-1.5 rounded-full px-3 py-1 font-display text-xs font-semibold shadow-soft',
            theme.badge,
          )}
        >
          <theme.Icon className="h-3.5 w-3.5" />
          {speciesLabel(pet.species)}
        </span>
        {sold ? (
          <span className="absolute right-3 top-3 rounded-full bg-ink/85 px-3 py-1 font-display text-xs font-semibold text-ink-fg">
            Sold
          </span>
        ) : null}
      </div>
      <div className="flex flex-1 flex-col p-5">
        <div className="flex items-baseline justify-between gap-2">
          <h3 className="font-display text-xl font-semibold">{pet.name}</h3>
          <span className="shrink-0 rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-semibold text-accent">
            {formatAge(pet.ageYears)}
          </span>
        </div>
        <p className="mt-2 line-clamp-2 text-sm text-muted">{pet.description}</p>
        <div data-testid="pet-card-actions" className="mt-5">
          {sold ? null : <PetCardActions pet={pet} refetch={refetch} />}
        </div>
      </div>
    </motion.li>
  );
}
