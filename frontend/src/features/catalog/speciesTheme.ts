import type { ComponentType } from 'react';
import { Species } from '@/__generated__/graphql';
import { CatIcon, DogIcon, FrogIcon } from '@/components/brand/AnimalIcons';

type IconComponent = ComponentType<{ className?: string }>;

interface SpeciesTheme {
  /** Friendly filled animal icon (currentColor). */
  Icon: IconComponent;
  /** Soft tinted background for image placeholders / tiles (Tailwind classes). */
  tile: string;
  /** Solid badge fill + readable text (Tailwind classes). */
  badge: string;
  /** Icon color on a light placeholder (Tailwind class). */
  ink: string;
}

const THEMES: Record<Species, SpeciesTheme> = {
  [Species.Cat]: { Icon: CatIcon, tile: 'bg-cat/15', badge: 'bg-cat text-ink', ink: 'text-cat' },
  [Species.Dog]: { Icon: DogIcon, tile: 'bg-dog/15', badge: 'bg-dog text-white', ink: 'text-dog' },
  [Species.Frog]: { Icon: FrogIcon, tile: 'bg-frog/15', badge: 'bg-frog text-white', ink: 'text-frog' },
};

export function speciesTheme(species: Species): SpeciesTheme {
  return THEMES[species];
}
