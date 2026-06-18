import { afterEach, describe, expect, it } from 'vitest';
import { credentialStore } from '../credentialStore';

afterEach(() => {
  credentialStore.clear();
  sessionStorage.clear();
  localStorage.clear();
});

describe('credentialStore', () => {
  it('stores and reads a credential', () => {
    credentialStore.set({ basic: 'YWJj', email: 'a@b.com' });
    expect(credentialStore.get()).toEqual({ basic: 'YWJj', email: 'a@b.com' });
  });

  it('mirrors to sessionStorage and never localStorage', () => {
    credentialStore.set({ basic: 'YWJj', email: 'a@b.com' });
    expect(sessionStorage.getItem('petstore.auth')).toContain('a@b.com');
    expect(localStorage.length).toBe(0);
  });

  it('clears memory and sessionStorage', () => {
    credentialStore.set({ basic: 'YWJj', email: 'a@b.com' });
    credentialStore.clear();
    expect(credentialStore.get()).toBeNull();
    expect(sessionStorage.getItem('petstore.auth')).toBeNull();
  });

  it('notifies subscribers on change', () => {
    let last: { basic: string; email: string } | null = null;
    const unsubscribe = credentialStore.subscribe((c) => {
      last = c;
    });
    credentialStore.set({ basic: 'x', email: 'e' });
    expect(last).toEqual({ basic: 'x', email: 'e' });
    unsubscribe();
  });
});
