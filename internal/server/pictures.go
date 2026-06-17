package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"roboticCrewChallenge/internal/domain"
)

// pictureCacheControl lets the browser and any shared cache hold a pet picture
// for a few minutes; object keys are immutable, so a stale image is never wrong.
const pictureCacheControl = "public, max-age=300"

// pictureReader is the slice of the object store the picture path needs: read a
// stored picture by its key. It is declared here so the server depends on the
// capability, not the adapter.
type pictureReader interface {
	Get(ctx context.Context, objectKey string) (domain.PictureContent, error)
}

// newPictureHandler streams a stored pet picture to the client over the same
// origin and TLS as the API, so the browser never talks to object storage
// directly and clients never see bucket keys or signed URLs. Pictures are public
// catalog content addressed by an opaque key, so the path is unauthenticated —
// the same access model a presigned URL would give, without exposing the bucket.
func newPictureHandler(pictures pictureReader, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		objectKey := r.PathValue("objectKey")
		if objectKey == "" {
			http.NotFound(w, r)
			return
		}

		content, err := pictures.Get(r.Context(), objectKey)
		if err != nil {
			if errors.Is(err, domain.ErrPictureNotFound) {
				http.NotFound(w, r)
				return
			}
			logger.ErrorContext(r.Context(), "serve picture", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		defer func() { _ = content.Body.Close() }()

		w.Header().Set("Content-Type", content.ContentType)
		// The content type is fixed at upload from sniffed bytes against an image
		// allowlist; nosniff stops a client from reinterpreting it as anything else.
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if content.Size >= 0 {
			w.Header().Set("Content-Length", strconv.FormatInt(content.Size, 10))
		}
		w.Header().Set("Cache-Control", pictureCacheControl)
		if _, err := io.Copy(w, content.Body); err != nil {
			logger.ErrorContext(r.Context(), "stream picture", "error", err)
		}
	})
}
