/**
 * Holds the signed-in customer's HTTP Basic credential for the order flow. In memory + mirrored
 * to sessionStorage (survives an in-tab refresh, dies on tab close) — NEVER localStorage. The
 * Apollo auth link reads it via get(); React mirrors it via subscribe().
 */
const STORAGE_KEY = 'petstore.auth';

export interface StoredCredential {
  /** base64("email:password") — the Authorization value without the "Basic " prefix. */
  basic: string;
  /** Plain email, for display. */
  email: string;
}

type Listener = (credential: StoredCredential | null) => void;

const listeners = new Set<Listener>();

function readFromSession(): StoredCredential | null {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    const parsed: unknown = JSON.parse(raw);
    if (
      typeof parsed === 'object' &&
      parsed !== null &&
      typeof (parsed as Record<string, unknown>).basic === 'string' &&
      typeof (parsed as Record<string, unknown>).email === 'string'
    ) {
      return parsed as StoredCredential;
    }
    return null;
  } catch {
    return null;
  }
}

let current: StoredCredential | null = readFromSession();

function notify(): void {
  for (const listener of listeners) listener(current);
}

export const credentialStore = {
  get(): StoredCredential | null {
    return current;
  },
  set(credential: StoredCredential): void {
    current = credential;
    try {
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify(credential));
    } catch {
      // sessionStorage unavailable — keep the credential in memory only.
    }
    notify();
  },
  clear(): void {
    current = null;
    try {
      sessionStorage.removeItem(STORAGE_KEY);
    } catch {
      // ignore
    }
    notify();
  },
  subscribe(listener: Listener): () => void {
    listeners.add(listener);
    return () => {
      listeners.delete(listener);
    };
  },
};
