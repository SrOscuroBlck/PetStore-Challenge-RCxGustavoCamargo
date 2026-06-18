# Frontend threat model

The challenge requires common web attack vectors to be addressed and the threat model documented. The
backend's model is in `docs/SECURITY.md`; this covers the **customer SPA**. Scope is the browser-side
trust boundary: what the app holds, what can attack it, and how each risk is mitigated — and, where
possible, mechanically enforced.

## Assets & trust boundaries

| Asset | Sensitivity | Where it lives |
|---|---|---|
| Ambient browse credential (`base64(email:password)`) | High | **Gateway-side only** (K8s Secret / dev `.env`); never in the browser (ADR-0005) |
| Signed-in customer credential | High | In memory + `sessionStorage` **only after the user signs in to order**, cleared on logout; never `localStorage` (ADR-0006) |
| In-flight GraphQL requests/responses | Medium | Network, TLS-protected |
| Catalog data / pet images | Low (public catalog) | Apollo cache / `<img>` |

Trust boundaries: **browser ↔ API** (same-origin over TLS, ADR-0003) and **app ↔ third-party code**
(npm dependencies, the runtime DOM).

## Vectors & mitigations

| Vector | Risk | Mitigation | Enforced by |
|---|---|---|---|
| **XSS** | Hijack the session / act as the user | React auto-escaping; **no `dangerouslySetInnerHTML` on server/user data**; a strict Content-Security-Policy; never `eval`/inject HTML. The credential isn't in the browser, so XSS cannot exfiltrate it | `guard.js` blocks `dangerouslySetInnerHTML`; CSP set at serve time (nginx) |
| **Credential exposure in the browser** | Credential theft | The customer credential lives **only at the gateway** (ADR-0005) — never in the bundle, JS, or storage. `localStorage` is also banned outright | `guard.js` blocks `localStorage`; ADR-0005 |
| **MITM / plaintext** | Credential or data interception | TLS only; same-origin behind the ingress; relative URLs (no mixed content); plaintext HTTP refused by backend | ADR-0003 |
| **Secrets in the bundle** | Leaking keys shipped to the client | No secrets in frontend code/env; only non-sensitive `VITE_*` build vars; `.env` git-ignored, documented via `.env.example` | `.gitignore`; `bash-guard.js` blocks `git add .env` |
| **Privilege confusion (calling merchant ops)** | `FORBIDDEN` errors / accidental surface | Customer-only app; merchant ops and breeder fields never referenced | `guard.js` + `contract-checker` agent |
| **Clickjacking** | UI redress to trigger purchases | `frame-ancestors 'none'` in CSP / `X-Frame-Options: DENY` at serve time | nginx headers (deployment phase) |
| **Vulnerable dependencies** | Supply-chain compromise of the bundle | Pinned lockfile; `npm audit` in the quality gate; minimal dependency surface | `quality-gate` skill |
| **Sensitive data in logs/errors** | Leaking credentials or PII | Never log the credential; show backend's human-readable `message`, never raw internals; `PublicPet` carries no breeder PII | code review |
| **Over-fetch / DoS-ish queries** | Expensive queries, slow UX | `first` ≤ 100, cursor pagination, normalized cache | `guard.js` blocks `first > 100` |
| **Untrusted image sources** | Loading attacker content | Images come only from the API's same-origin `/pictures/{key}` path; graceful placeholder on 404 | code review |

## Residual risk & assumptions

- **Self-signed TLS in local dev** must be trusted once (documented in the README). Not a production
  posture; acceptable for the local-only challenge.
- **Open storefront** (ADR-0005): anyone who can reach the host can browse/purchase as the single demo
  customer — identical to a public storefront, and acceptable for a money-less local demo. Endpoints
  remain auth-gated; a direct unauthenticated API call is still rejected.
- The `/pictures` path is intentionally unauthenticated public catalog content (backend decision,
  [`../../docs/adr/0007-picture-proxy-path.md`](../../docs/adr/0007-picture-proxy-path.md)) — no customer PII is served there.

## Content-Security-Policy (target)

Set by the serving nginx in the deployment phase; baseline intent:

```
default-src 'self';
img-src 'self' data:;
script-src 'self';
style-src 'self' 'unsafe-inline';   /* Tailwind-generated styles; tighten with a nonce if feasible */
connect-src 'self';
frame-ancestors 'none';
base-uri 'self';
object-src 'none';
```
