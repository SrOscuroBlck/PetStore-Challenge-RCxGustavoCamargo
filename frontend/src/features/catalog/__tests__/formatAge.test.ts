import { describe, expect, it } from 'vitest';
import { Species } from '@/__generated__/graphql';
import { formatAge } from '../formatAge';
import { speciesLabel } from '../speciesLabel';

describe('formatAge', () => {
  it('uses the singular for one year', () => {
    expect(formatAge(1)).toBe('1 yr');
  });
  it('uses the plural otherwise', () => {
    expect(formatAge(0)).toBe('0 yrs');
    expect(formatAge(3)).toBe('3 yrs');
  });
});

describe('speciesLabel', () => {
  it('maps each enum value to a display label', () => {
    expect(speciesLabel(Species.Cat)).toBe('Cat');
    expect(speciesLabel(Species.Dog)).toBe('Dog');
    expect(speciesLabel(Species.Frog)).toBe('Frog');
  });
});
