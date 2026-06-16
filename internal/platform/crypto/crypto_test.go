package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return key
}

func TestEncryptor_RoundTrip(t *testing.T) {
	enc, err := NewEncryptor(testKey(t))
	if err != nil {
		t.Fatalf("new encryptor: %v", err)
	}

	plaintext := []byte("breeder@example.com")
	ciphertext, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext must not equal plaintext")
	}

	got, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("round trip mismatch: got %q want %q", got, plaintext)
	}
}

func TestNewEncryptor_RejectsWrongKeySize(t *testing.T) {
	if _, err := NewEncryptor([]byte("too-short")); err == nil {
		t.Fatal("expected an error for a key that is not 32 bytes")
	}
}

func TestPassword_HashThenVerify(t *testing.T) {
	hash, err := HashPassword("s3cret-password")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if err := VerifyPassword(hash, "s3cret-password"); err != nil {
		t.Fatalf("expected the correct password to verify, got %v", err)
	}
	if err := VerifyPassword(hash, "wrong-password"); err == nil {
		t.Fatal("expected a wrong password to fail verification")
	}
}

func TestBlindIndex_DeterministicAndNormalised(t *testing.T) {
	index, err := NewBlindIndex(testKey(t))
	if err != nil {
		t.Fatalf("new blind index: %v", err)
	}

	mixedCase := index.Compute("User@Example.com")
	normalised := index.Compute("  user@example.com ")
	if !bytes.Equal(mixedCase, normalised) {
		t.Fatal("expected normalised inputs to produce the same index")
	}

	different := index.Compute("someone-else@example.com")
	if bytes.Equal(mixedCase, different) {
		t.Fatal("expected different inputs to produce different indexes")
	}
}
