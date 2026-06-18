import { ApolloError, type ServerError, type ServerParseError } from '@apollo/client';
import { isErrorCode, type ErrorCode } from './errorCodes';

export interface MappedError {
  code: ErrorCode | 'NETWORK' | 'UNKNOWN';
  /** Human-readable, safe to show the user. */
  userMessage: string;
}

const GENERIC = 'Something went wrong. Please try again.';

/**
 * Pick the user-facing message for a code. For UNAVAILABLE we prefer the server's message
 * because it is human-readable and, for checkout, names the unavailable pets (challenge req).
 * Internal/validation/build-time codes get a safe generic so we never leak details.
 */
function messageForCode(code: ErrorCode, serverMessage: string): string {
  switch (code) {
    case 'UNAVAILABLE':
      return serverMessage || 'That pet is no longer available.';
    case 'NOT_FOUND':
      return 'This pet is no longer listed.';
    case 'UNAUTHENTICATED':
      return 'Your session has ended. Please sign in again.';
    case 'FORBIDDEN':
      return 'You do not have access to that.';
    case 'CONFLICT':
    case 'VALIDATION':
    case 'GRAPHQL_VALIDATION_FAILED':
    case 'COMPLEXITY_LIMIT_EXCEEDED':
    case 'UNSUPPORTED_MEDIA_TYPE':
    case 'PAYLOAD_TOO_LARGE':
    case 'INTERNAL':
      return GENERIC;
  }
}

function statusCodeOf(networkError: Error | ServerError | ServerParseError | null): number | undefined {
  if (networkError && 'statusCode' in networkError) {
    return (networkError as { statusCode?: number }).statusCode;
  }
  return undefined;
}

/** Map any thrown error from an operation to a code + a user-facing message. */
export function mapApolloError(error: unknown): MappedError {
  if (error instanceof ApolloError) {
    for (const gqlError of error.graphQLErrors) {
      const code = gqlError.extensions?.code;
      if (isErrorCode(code)) {
        return { code, userMessage: messageForCode(code, gqlError.message) };
      }
    }
    if (error.networkError) {
      return { code: 'NETWORK', userMessage: 'Network error. Check your connection and try again.' };
    }
  }
  return { code: 'UNKNOWN', userMessage: GENERIC };
}

/** True when the error means bad/missing credentials — drives the login screen. */
export function isUnauthenticatedError(error: unknown): boolean {
  if (!(error instanceof ApolloError)) return false;
  if (error.graphQLErrors.some((e) => e.extensions?.code === 'UNAUTHENTICATED')) return true;
  return statusCodeOf(error.networkError) === 401;
}
