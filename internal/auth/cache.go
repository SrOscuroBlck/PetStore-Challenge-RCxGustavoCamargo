package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// credentialCache memoizes successful authentications for a short window so the
// deliberately-expensive bcrypt verification does not run on every request — the
// dominant cost that otherwise caps throughput far below the 1k-concurrent-user
// target (see docs/PERFORMANCE.md). The key is a hash of the credential, so only
// a caller presenting the correct password can reach a cached entry; a revoked or
// changed credential keeps working for at most the TTL. Only successes are cached,
// and the seeded account set is tiny, so the entry count is naturally bounded.
type credentialCache struct {
	ttl     time.Duration
	now     func() time.Time
	mu      sync.RWMutex
	entries map[string]cacheEntry
}

type cacheEntry struct {
	identity  Identity
	expiresAt time.Time
}

func newCredentialCache(ttl time.Duration) *credentialCache {
	return &credentialCache{
		ttl:     ttl,
		now:     time.Now,
		entries: make(map[string]cacheEntry),
	}
}

// credentialKey hashes email+password so the plaintext password is never used as
// a map key. The null separator prevents distinct (email, password) pairs from
// colliding by concatenation.
func credentialKey(email, password string) string {
	sum := sha256.Sum256([]byte(email + "\x00" + password))
	return hex.EncodeToString(sum[:])
}

func (c *credentialCache) get(key string) (Identity, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok || !c.now().Before(entry.expiresAt) {
		return Identity{}, false
	}
	return entry.identity, true
}

func (c *credentialCache) put(key string, identity Identity) {
	c.mu.Lock()
	c.entries[key] = cacheEntry{identity: identity, expiresAt: c.now().Add(c.ttl)}
	c.mu.Unlock()
}
