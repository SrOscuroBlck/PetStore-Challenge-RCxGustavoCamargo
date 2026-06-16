package pagination

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCursor_RoundTrip(t *testing.T) {
	want := Cursor{
		SortKey: time.Now().UTC().Truncate(time.Microsecond),
		ID:      uuid.New(),
	}

	got, err := Decode(Encode(want))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !got.SortKey.Equal(want.SortKey) {
		t.Fatalf("sortKey mismatch: got %v want %v", got.SortKey, want.SortKey)
	}
	if got.ID != want.ID {
		t.Fatalf("id mismatch: got %v want %v", got.ID, want.ID)
	}
}

func TestDecode_RejectsGarbage(t *testing.T) {
	if _, err := Decode("!!!not-base64!!!"); err == nil {
		t.Fatal("expected an error decoding an invalid cursor")
	}
}
