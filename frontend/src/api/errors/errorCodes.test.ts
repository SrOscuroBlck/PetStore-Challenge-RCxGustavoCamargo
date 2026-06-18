import { describe, expect, it } from 'vitest';
import { isErrorCode } from './errorCodes';

// Smoke test confirming the Vitest harness runs. Critical-flow tests land in issue #6.
describe('isErrorCode', () => {
  it('accepts known backend error codes', () => {
    expect(isErrorCode('UNAVAILABLE')).toBe(true);
    expect(isErrorCode('UNAUTHENTICATED')).toBe(true);
  });

  it('rejects unknown values', () => {
    expect(isErrorCode('NOPE')).toBe(false);
    expect(isErrorCode(123)).toBe(false);
    expect(isErrorCode(null)).toBe(false);
  });
});
