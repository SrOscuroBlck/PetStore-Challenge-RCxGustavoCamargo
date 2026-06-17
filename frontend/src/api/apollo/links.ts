import { HttpLink } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';
import { credentialStore } from '@/features/auth/credentialStore';

/** Relative URI — same-origin in prod, proxied in dev. The gateway injects the ambient
    browse credential; when signed in, we attach the customer's own credential below. */
export const httpLink = new HttpLink({ uri: '/graphql' });

/**
 * Attach the signed-in customer's HTTP Basic to requests (login-to-order, ADR-0006). When
 * signed out, send no Authorization so the gateway injects the ambient browse credential. A
 * per-request Authorization (the login probe) is respected and takes precedence.
 */
export const authLink = setContext((_operation, prevContext) => {
  const headers = (prevContext.headers ?? {}) as Record<string, string>;
  if (headers.Authorization || headers.authorization) return { headers };
  const credential = credentialStore.get();
  if (!credential) return { headers };
  return { headers: { ...headers, Authorization: `Basic ${credential.basic}` } };
});
