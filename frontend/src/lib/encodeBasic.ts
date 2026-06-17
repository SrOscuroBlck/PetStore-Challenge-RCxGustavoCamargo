/**
 * Build the HTTP Basic credential value: base64("email:password").
 * `btoa` only accepts Latin-1, so we UTF-8 encode first — a non-ASCII password would
 * otherwise throw (emails are ASCII per the backend, but passwords may not be).
 */
export function encodeBasic(email: string, password: string): string {
  const bytes = new TextEncoder().encode(`${email}:${password}`);
  let binary = '';
  for (const byte of bytes) binary += String.fromCharCode(byte);
  return btoa(binary);
}
