import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from 'react';
import { useApolloClient } from '@apollo/client';
import { AvailablePetsDocument } from '@/__generated__/graphql';
import { mapApolloError, isUnauthenticatedError } from '@/api/errors/mapError';
import { encodeBasic } from '@/lib/encodeBasic';
import { credentialStore } from './credentialStore';
import { AuthContext, type AuthApi, type AuthStatus, type LoginResult } from './authContext';

/**
 * Open browse, login-to-order (ADR-0006). Browsing needs no sign-in (the gateway injects an
 * ambient credential); placing an order opens the login dialog via `ensureSignedIn()`, and the
 * stored credential is then attached to requests (the gateway passes it through).
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const client = useApolloClient();
  const [credential, setCredential] = useState(() => credentialStore.get());
  const [status, setStatus] = useState<AuthStatus>(() =>
    credentialStore.get() ? 'signedIn' : 'signedOut',
  );
  const [isLoginOpen, setIsLoginOpen] = useState(false);
  const resolverRef = useRef<((ok: boolean) => void) | null>(null);

  useEffect(
    () =>
      credentialStore.subscribe((next) => {
        setCredential(next);
        setStatus(next ? 'signedIn' : 'signedOut');
      }),
    [],
  );

  const settle = useCallback((ok: boolean) => {
    const resolve = resolverRef.current;
    resolverRef.current = null;
    resolve?.(ok);
  }, []);

  const ensureSignedIn = useCallback((): Promise<boolean> => {
    if (credentialStore.get()) return Promise.resolve(true);
    setIsLoginOpen(true);
    return new Promise<boolean>((resolve) => {
      resolverRef.current = resolve;
    });
  }, []);

  const closeLogin = useCallback(() => {
    setIsLoginOpen(false);
    settle(false);
  }, [settle]);

  const login = useCallback<AuthApi['login']>(
    async (email, password, storeId): Promise<LoginResult> => {
      setStatus('authenticating');
      const basic = encodeBasic(email, password);
      try {
        // Probe with the candidate credential before storing it. An empty connection (unknown
        // store) still means the credentials are valid.
        await client.query({
          query: AvailablePetsDocument,
          variables: { storeId, species: null, first: 1 },
          fetchPolicy: 'network-only',
          context: { headers: { Authorization: `Basic ${basic}` } },
        });
        credentialStore.set({ basic, email });
        setIsLoginOpen(false);
        settle(true);
        return { ok: true };
      } catch (error) {
        setStatus(credentialStore.get() ? 'signedIn' : 'signedOut');
        if (isUnauthenticatedError(error)) {
          return { ok: false, errorMessage: 'Invalid email or password.' };
        }
        return { ok: false, errorMessage: mapApolloError(error).userMessage };
      }
    },
    [client, settle],
  );

  const logout = useCallback(() => {
    credentialStore.clear();
    void client.clearStore();
  }, [client]);

  const value = useMemo<AuthApi>(
    () => ({
      status,
      email: credential?.email ?? null,
      isSignedIn: credential !== null,
      isLoginOpen,
      ensureSignedIn,
      login,
      logout,
      closeLogin,
    }),
    [status, credential, isLoginOpen, ensureSignedIn, login, logout, closeLogin],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
