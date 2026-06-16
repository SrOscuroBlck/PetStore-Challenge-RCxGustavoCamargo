package picture

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"roboticCrewChallenge/internal/domain"
)

var (
	pngHeader  = []byte("\x89PNG\r\n\x1a\n")
	jpegHeader = []byte("\xff\xd8\xff\xe0")
	webpHeader = []byte("RIFF\x00\x00\x00\x00WEBPVP")
)

func TestValidate_AcceptsAllowedTypes(t *testing.T) {
	cases := map[string]struct {
		body []byte
		want string
	}{
		"png":  {append(pngHeader, []byte("more pixel data")...), "image/png"},
		"jpeg": {append(jpegHeader, []byte("more pixel data")...), "image/jpeg"},
		"webp": {append(webpHeader, []byte("more pixel data")...), "image/webp"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ct, size, r, err := Validate(bytes.NewReader(tc.body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ct != tc.want {
				t.Fatalf("content type = %q, want %q", ct, tc.want)
			}
			if size != int64(len(tc.body)) {
				t.Fatalf("size = %d, want %d", size, len(tc.body))
			}
			got, err := io.ReadAll(r)
			if err != nil {
				t.Fatalf("read returned reader: %v", err)
			}
			if !bytes.Equal(got, tc.body) {
				t.Fatal("returned reader did not replay the full original bytes")
			}
		})
	}
}

func TestValidate_RejectsUnsupportedType(t *testing.T) {
	_, _, _, err := Validate(bytes.NewReader([]byte("just some plain text, definitely not an image")))
	if !errors.Is(err, domain.ErrUnsupportedPictureType) {
		t.Fatalf("expected ErrUnsupportedPictureType, got %v", err)
	}
}

func TestValidate_RejectsEmpty(t *testing.T) {
	_, _, _, err := Validate(bytes.NewReader(nil))
	if !errors.Is(err, domain.ErrUnsupportedPictureType) {
		t.Fatalf("expected ErrUnsupportedPictureType for empty input, got %v", err)
	}
}

func TestValidate_RejectsOversize(t *testing.T) {
	body := make([]byte, MaxPictureBytes+1)
	copy(body, pngHeader)

	_, _, _, err := Validate(bytes.NewReader(body))
	if !errors.Is(err, domain.ErrPictureTooLarge) {
		t.Fatalf("expected ErrPictureTooLarge, got %v", err)
	}
}

func TestValidate_AcceptsExactlyMaxSize(t *testing.T) {
	body := make([]byte, MaxPictureBytes)
	copy(body, pngHeader)

	_, size, _, err := Validate(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error at the size ceiling: %v", err)
	}
	if size != MaxPictureBytes {
		t.Fatalf("size = %d, want %d", size, MaxPictureBytes)
	}
}
