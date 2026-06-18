import { useContext } from 'react';
import { CartContext, type CartApi } from './cartContext';

export function useCart(): CartApi {
  const ctx = useContext(CartContext);
  if (!ctx) throw new Error('useCart must be used within <CartProvider>');
  return ctx;
}
