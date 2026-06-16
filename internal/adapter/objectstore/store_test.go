package objectstore

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/platform/picture"
)

var pngBytes = append([]byte("\x89PNG\r\n\x1a\n"), []byte("fake-but-sniffable-png-pixel-data")...)

func TestUpload_StoresObjectAndReturnsKey(t *testing.T) {
	requireStore(t)
	ctx := context.Background()

	key, err := testStore.Upload(ctx, bytes.NewReader(pngBytes), int64(len(pngBytes)), "image/png")
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if !strings.HasPrefix(key, "pets/") {
		t.Fatalf("object key %q should be namespaced under pets/", key)
	}
}

func TestPresignedURL_ServesUploadedBytes(t *testing.T) {
	requireStore(t)
	ctx := context.Background()

	key, err := testStore.Upload(ctx, bytes.NewReader(pngBytes), int64(len(pngBytes)), "image/png")
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	url, err := testStore.PresignedURL(ctx, key)
	if err != nil {
		t.Fatalf("presign: %v", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("fetch presigned url: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("presigned GET status = %d, want 200", resp.StatusCode)
	}
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !bytes.Equal(got, pngBytes) {
		t.Fatal("presigned URL did not serve the uploaded bytes")
	}
}

// The picture mutation in #10 validates before uploading; this proves bad input
// is rejected with no reader to stream, so nothing can reach the store.
func TestValidateThenUpload_RejectsBadInput(t *testing.T) {
	t.Run("oversize", func(t *testing.T) {
		body := make([]byte, picture.MaxPictureBytes+1)
		copy(body, pngBytes)
		assertValidationRejects(t, bytes.NewReader(body), domain.ErrPictureTooLarge)
	})

	t.Run("disallowed type", func(t *testing.T) {
		body := []byte("plain text masquerading as a pet picture")
		assertValidationRejects(t, bytes.NewReader(body), domain.ErrUnsupportedPictureType)
	})
}

func assertValidationRejects(t *testing.T, body io.Reader, want error) {
	t.Helper()
	_, _, validated, err := picture.Validate(body)
	if !errors.Is(err, want) {
		t.Fatalf("validate error = %v, want %v", err, want)
	}
	if validated != nil {
		t.Fatal("rejected input must not yield a reader to upload")
	}
}
