import { describe, expect, it } from 'vitest';
import { Species } from '@/__generated__/graphql';
import { cartReducer } from '../cartReducer';
import type { CartItem } from '../cartContext';

const cat: CartItem = { id: '1', name: 'Whiskers', species: Species.Cat, pictureUrl: '/p/1' };
const dog: CartItem = { id: '2', name: 'Rex', species: Species.Dog, pictureUrl: '/p/2' };

describe('cartReducer', () => {
  it('adds an item', () => {
    expect(cartReducer([], { type: 'add', item: cat })).toEqual([cat]);
  });

  it('dedupes by id — a pet cannot be added twice', () => {
    const once = cartReducer([], { type: 'add', item: cat });
    const twice = cartReducer(once, { type: 'add', item: cat });
    expect(twice).toHaveLength(1);
    expect(twice).toBe(once); // unchanged reference
  });

  it('removes by id', () => {
    expect(cartReducer([cat, dog], { type: 'remove', id: '1' })).toEqual([dog]);
  });

  it('clears', () => {
    expect(cartReducer([cat, dog], { type: 'clear' })).toEqual([]);
  });
});
