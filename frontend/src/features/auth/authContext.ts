import { createContext } from 'react';

export type AuthStatus = 'signedOut' | 'authenticating' | 'signedIn';

export interface LoginResult {
  ok: boolean;
  errorMessage?: string;
}

export interface AuthApi {
  status: AuthStatus;
  email: string | null;
  isSignedIn: boolean;
  isLoginOpen: boolean;
  /** Open the login dialog if signed out; resolves true once signed in, false if cancelled. */
  ensureSignedIn: () => Promise<boolean>;
  /** Validate + store credentials (login-by-probe against the routed store). */
  login: (email: string, password: string, storeId: string) => Promise<LoginResult>;
  logout: () => void;
  closeLogin: () => void;
}

export const AuthContext = createContext<AuthApi | null>(null);
