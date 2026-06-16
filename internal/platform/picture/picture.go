package picture

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"roboticCrewChallenge/internal/domain"
)

// MaxPictureBytes caps a pet picture upload before it is streamed to object
// storage, guarding against memory and storage abuse.
const MaxPictureBytes int64 = 5 << 20

var allowedContentTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
}

// Validate enforces the size cap and content-type allowlist for a pet picture.
// On success it returns a fresh reader replaying the full content together with
// its exact size, so the caller can stream it to storage in one shot; memory is
// bounded by MaxPictureBytes. On rejection the returned reader is nil.
func Validate(body io.Reader) (string, int64, io.Reader, error) {
	buf, err := io.ReadAll(io.LimitReader(body, MaxPictureBytes+1))
	if err != nil {
		return "", 0, nil, fmt.Errorf("read picture: %w", err)
	}
	if int64(len(buf)) > MaxPictureBytes {
		return "", 0, nil, domain.ErrPictureTooLarge
	}

	contentType := http.DetectContentType(buf)
	if _, ok := allowedContentTypes[contentType]; !ok {
		return "", 0, nil, domain.ErrUnsupportedPictureType
	}

	return contentType, int64(len(buf)), bytes.NewReader(buf), nil
}
