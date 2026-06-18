package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCredentialCache_ServesHitWithinTTL(t *testing.T) {
	clock := time.Unix(0, 0)
	cache := newCredentialCache(60 * time.Second)
	cache.now = func() time.Time { return clock }

	key := credentialKey("customer@petstore.local", "demo-password")
	want := Identity{Subject: uuid.New(), Role: RoleCustomer}
	cache.put(key, want)

	clock = clock.Add(59 * time.Second)
	got, ok := cache.get(key)
	if !ok {
		t.Fatal("expected a cache hit within the TTL")
	}
	if got != want {
		t.Fatalf("cached identity = %+v, want %+v", got, want)
	}
}

func TestCredentialCache_ExpiresAfterTTL(t *testing.T) {
	clock := time.Unix(0, 0)
	cache := newCredentialCache(60 * time.Second)
	cache.now = func() time.Time { return clock }

	key := credentialKey("customer@petstore.local", "demo-password")
	cache.put(key, Identity{Subject: uuid.New(), Role: RoleCustomer})

	clock = clock.Add(60 * time.Second)
	if _, ok := cache.get(key); ok {
		t.Fatal("expected the entry to expire exactly at the TTL boundary")
	}
}

func TestCredentialKey_DependsOnBothFields(t *testing.T) {
	base := credentialKey("a@petstore.local", "secret")
	if base == credentialKey("b@petstore.local", "secret") {
		t.Fatal("key must change when the email changes")
	}
	if base == credentialKey("a@petstore.local", "secret2") {
		t.Fatal("key must change when the password changes")
	}
	// A wrong password must never resolve to a cached success for the right one.
	if base == credentialKey("a@petstore.local\x00secret", "") {
		t.Fatal("separator must prevent field-concatenation collisions")
	}
}
