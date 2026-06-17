import { useCallback, useEffect, useMemo, useReducer, useState, type ReactNode } from 'react';
import { CartContext, type CartApi, type CartItem } from './cartContext';
import { cartReducer } from './cartReducer';

const STORAGE_KEY = 'petstore.cart';

// Hydrate from sessionStorage so an in-tab refresh keeps the cart (never localStorage).
function hydrate(): CartItem[] {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter((item): item is CartItem => {
      if (typeof item !== 'object' || item === null) return false;
      const record = item as Record<string, unknown>;
      return (
        typeof record.id === 'string' &&
        typeof record.name === 'string' &&
        typeof record.species === 'string' &&
        typeof record.pictureUrl === 'string'
      );
    });
  } catch {
    return [];
  }
}

export function CartProvider({ children }: { children: ReactNode }) {
  const [items, dispatch] = useReducer(cartReducer, undefined, hydrate);
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    try {
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify(items));
    } catch {
      // sessionStorage unavailable (private mode) — cart stays in memory.
    }
  }, [items]);

  const add = useCallback((item: CartItem) => dispatch({ type: 'add', item }), []);
  const remove = useCallback((id: string) => dispatch({ type: 'remove', id }), []);
  const clear = useCallback(() => dispatch({ type: 'clear' }), []);
  const has = useCallback((id: string) => items.some((item) => item.id === id), [items]);
  const open = useCallback(() => setIsOpen(true), []);
  const close = useCallback(() => setIsOpen(false), []);

  const value = useMemo<CartApi>(
    () => ({ items, count: items.length, has, add, remove, clear, isOpen, open, close }),
    [items, has, add, remove, clear, isOpen, open, close],
  );

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
}
