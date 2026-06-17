// Package rediscache implements the domain.CatalogCache port over Redis using
// cache-aside with per-store generation keys for O(1) invalidation and
// AES-256-GCM-encrypted payloads so breeder PII never sits in plaintext. It is
// non-fatal: any Redis error degrades to a cache miss or a no-op so callers
// always fall back to Postgres.
package rediscache
