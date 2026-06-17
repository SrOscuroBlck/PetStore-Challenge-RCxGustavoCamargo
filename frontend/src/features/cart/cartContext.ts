import { createContext } from 'react';
import type { Species } from '@/__generated__/graphql';

/** The minimum a cart line item needs to render — a narrow projection of a catalog pet. */
export interface CartItem {
  id: string;
  name: string;
  species: Species;
  pictureUrl: string;
}

export interface CartApi {
  items: CartItem[];
  count: number;
  has: (id: string) => boolean;
  add: (item: CartItem) => void;
  remove: (id: string) => void;
  clear: () => void;
  /** Drawer open-state lives here so the header button and the drawer coordinate. */
  isOpen: boolean;
  open: () => void;
  close: () => void;
}

export const CartContext = createContext<CartApi | null>(null);
