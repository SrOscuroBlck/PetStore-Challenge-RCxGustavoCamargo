import { Species } from '@/__generated__/graphql';

const LABELS: Record<Species, string> = {
  [Species.Cat]: 'Cat',
  [Species.Dog]: 'Dog',
  [Species.Frog]: 'Frog',
};

/** Display label for a species enum value (`CAT` → "Cat"). */
export function speciesLabel(species: Species): string {
  return LABELS[species];
}
