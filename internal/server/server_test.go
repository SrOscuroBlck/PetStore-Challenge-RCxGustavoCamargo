package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"roboticCrewChallenge/internal/auth"
	"roboticCrewChallenge/internal/config"
	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/platform/tlscert"
)

type fakePictures struct {
	content domain.PictureContent
	err     error
}

func (f fakePictures) Get(context.Context, string) (domain.PictureContent, error) {
	return f.content, f.err
}

func selfSignedFiles(t *testing.T) (certFile, keyFile string) {
	t.Helper()
	certPEM, keyPEM, err := tlscert.Generate([]string{"localhost", "127.0.0.1"})
	if err != nil {
		t.Fatalf("generate cert: %v", err)
	}
	dir := t.TempDir()
	certFile = filepath.Join(dir, "cert.pem")
	keyFile = filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certFile, certPEM, 0o600); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}
	return certFile, keyFile
}

func TestHandleHealth_ReturnsOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handleHealth(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected an ok status body, got %q", rec.Body.String())
	}
}

func TestServer_PlaygroundServedUnauthenticatedWhenProvided(t *testing.T) {
	certFile, keyFile := selfSignedFiles(t)
	cfg := config.Config{HTTPAddr: "127.0.0.1:0", TLSCertFile: certFile, TLSKeyFile: keyFile}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	stub := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	playgroundStub := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "PLAYGROUND")
	})
	srv := New(cfg, logger, auth.NewAuthenticator(nil, nil, nil), stub, playgroundStub, nil)

	req := httptest.NewRequest(http.MethodGet, "/playground", nil)
	rec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("playground status = %d, want 200 with no credentials", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "PLAYGROUND") {
		t.Fatalf("playground body = %q, want the playground page", rec.Body.String())
	}
}

func TestServer_PlaygroundAbsentWhenNil(t *testing.T) {
	certFile, keyFile := selfSignedFiles(t)
	cfg := config.Config{HTTPAddr: "127.0.0.1:0", TLSCertFile: certFile, TLSKeyFile: keyFile}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	stub := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	srv := New(cfg, logger, auth.NewAuthenticator(nil, nil, nil), stub, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/playground", nil)
	rec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("playground status = %d, want 404 when not mounted", rec.Code)
	}
}

func TestServer_PictureStreamedWhenProvided(t *testing.T) {
	certFile, keyFile := selfSignedFiles(t)
	cfg := config.Config{HTTPAddr: "127.0.0.1:0", TLSCertFile: certFile, TLSKeyFile: keyFile}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	stub := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	pics := fakePictures{content: domain.PictureContent{
		Body:        io.NopCloser(strings.NewReader("IMG-BYTES")),
		ContentType: "image/png",
		Size:        9,
	}}
	srv := New(cfg, logger, auth.NewAuthenticator(nil, nil, nil), stub, nil, pics)

	req := httptest.NewRequest(http.MethodGet, "/pictures/pets/abc", nil)
	rec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("picture status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "IMG-BYTES" {
		t.Fatalf("picture body = %q, want the streamed bytes", rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("content type = %q, want image/png", ct)
	}
}

func TestServer_PictureNotFoundMapsTo404(t *testing.T) {
	certFile, keyFile := selfSignedFiles(t)
	cfg := config.Config{HTTPAddr: "127.0.0.1:0", TLSCertFile: certFile, TLSKeyFile: keyFile}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	stub := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	pics := fakePictures{err: domain.ErrPictureNotFound}
	srv := New(cfg, logger, auth.NewAuthenticator(nil, nil, nil), stub, nil, pics)

	req := httptest.NewRequest(http.MethodGet, "/pictures/pets/missing", nil)
	rec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing picture status = %d, want 404", rec.Code)
	}
}

func TestServer_PictureAbsentWhenNil(t *testing.T) {
	certFile, keyFile := selfSignedFiles(t)
	cfg := config.Config{HTTPAddr: "127.0.0.1:0", TLSCertFile: certFile, TLSKeyFile: keyFile}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	stub := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	srv := New(cfg, logger, auth.NewAuthenticator(nil, nil, nil), stub, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/pictures/pets/abc", nil)
	rec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("picture status = %d, want 404 when not mounted", rec.Code)
	}
}

func TestServer_RunShutsDownWhenContextCancelled(t *testing.T) {
	certFile, keyFile := selfSignedFiles(t)
	cfg := config.Config{HTTPAddr: "127.0.0.1:0", LogLevel: slog.LevelInfo, TLSCertFile: certFile, TLSKeyFile: keyFile}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	stub := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	srv := New(cfg, logger, auth.NewAuthenticator(nil, nil, nil), stub, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.Run(ctx) }()

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected a clean shutdown, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after the context was cancelled")
	}
}
