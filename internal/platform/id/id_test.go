package id

import "testing"

func TestNew_ReturnsDistinctVersion7UUIDs(t *testing.T) {
	a, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	b, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if a == b {
		t.Fatal("expected New to return distinct UUIDs")
	}
	if a.Version() != 7 {
		t.Fatalf("expected a UUIDv7, got version %d", a.Version())
	}
}
