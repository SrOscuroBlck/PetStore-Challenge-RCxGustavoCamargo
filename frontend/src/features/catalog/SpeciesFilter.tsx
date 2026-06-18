import type { ComponentType } from 'react';
import { Species } from '@/__generated__/graphql';
import { cn } from '@/lib/cn';
import { Paw } from '@/components/brand/Paw';
import { speciesLabel } from './speciesLabel';
import { speciesTheme } from './speciesTheme';

interface SpeciesFilterProps {
  value: Species | null;
  onChange: (species: Species | null) => void;
}

type IconComponent = ComponentType<{ className?: string }>;

// `null` = All (the backend returns every species when the arg is absent).
const OPTIONS: ReadonlyArray<{ label: string; Icon: IconComponent; value: Species | null }> = [
  { label: 'All', Icon: Paw, value: null },
  { label: speciesLabel(Species.Cat), Icon: speciesTheme(Species.Cat).Icon, value: Species.Cat },
  { label: speciesLabel(Species.Dog), Icon: speciesTheme(Species.Dog).Icon, value: Species.Dog },
  { label: speciesLabel(Species.Frog), Icon: speciesTheme(Species.Frog).Icon, value: Species.Frog },
];

/** Accessible toggle group driving the server-side species filter. */
export function SpeciesFilter({ value, onChange }: SpeciesFilterProps) {
  return (
    <div role="group" aria-label="Filter by species" className="flex flex-wrap gap-2.5">
      {OPTIONS.map(({ label, Icon, value: optValue }) => {
        const active = optValue === value;
        return (
          <button
            key={label}
            type="button"
            aria-pressed={active}
            onClick={() => onChange(optValue)}
            className={cn(
              'inline-flex items-center gap-2 rounded-full border-2 px-5 py-2 font-display text-sm font-semibold tracking-wide transition-all motion-safe:active:scale-95',
              active
                ? 'border-primary bg-primary text-primary-fg shadow-soft'
                : 'border-border bg-card text-fg hover:border-primary hover:bg-primary/5',
            )}
          >
            <Icon className="h-4 w-4" />
            {label}
          </button>
        );
      })}
    </div>
  );
}
