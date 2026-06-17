/** Typed, centralized access to the (non-sensitive) build-time env. */
export const env = {
  /** Optional dev convenience: redirect bare "/" to this store. */
  defaultStoreId: import.meta.env.VITE_DEFAULT_STORE_ID?.trim() || undefined,
} as const;
