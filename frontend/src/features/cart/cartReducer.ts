import type { CartItem } from './cartContext';

export type CartAction =
  | { type: 'add'; item: CartItem }
  | { type: 'remove'; id: string }
  | { type: 'clear' };

/** Pure cart state transitions. `add` dedupes by id (a pet can't be added twice). */
export function cartReducer(state: CartItem[], action: CartAction): CartItem[] {
  switch (action.type) {
    case 'add':
      return state.some((item) => item.id === action.item.id) ? state : [...state, action.item];
    case 'remove':
      return state.filter((item) => item.id !== action.id);
    case 'clear':
      return [];
  }
}
