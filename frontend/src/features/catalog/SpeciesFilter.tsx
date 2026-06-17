import { Species } from '@/__generated__/graphql';
import { cn } from '@/lib/cn';
import { speciesLabel } from './speciesLabel';

interface SpeciesFilterProps {
  value: Species | null;
  onChange: (species: Species | null) => void;
}

// `null` = All (the backend returns every species when the arg is absent).
const OPTIONS: ReadonlyArray<{ label: string; value: Species | null }> = [
  { label: 'All', value: null },
  { label: speciesLabel(Species.Cat), value: Species.Cat },
  { label: speciesLabel(Species.Dog), value: Species.Dog },
  { label: speciesLabel(Species.Frog), value: Species.Frog },
];

/** Accessible toggle group driving the server-side species filter. */
export function SpeciesFilter({ value, onChange }: SpeciesFilterProps) {
  return (
    <div role="group" aria-label="Filter by species" className="flex flex-wrap gap-2">
      {OPTIONS.map((opt) => {
        const active = opt.value === value;
        return (
          <button
            key={opt.label}
            type="button"
            aria-pressed={active}
            onClick={() => onChange(opt.value)}
            className={cn(
              'rounded-sm border px-4 py-2 font-display text-sm font-semibold uppercase tracking-wide transition-colors',
              active
                ? 'border-primary bg-primary text-primary-fg'
                : 'border-border bg-card text-fg hover:border-primary',
            )}
          >
            {opt.label}
          </button>
        );
      })}
    </div>
  );
}
