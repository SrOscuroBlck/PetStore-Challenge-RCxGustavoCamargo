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
	"roboticCrewChallenge/internal/platform/tlscert"
)

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

func TestServer_RunShutsDownWhenContextCancelled(t *testing.T) {
	certFile, keyFile := selfSignedFiles(t)
	cfg := config.Config{HTTPAddr: "127.0.0.1:0", LogLevel: slog.LevelInfo, TLSCertFile: certFile, TLSKeyFile: keyFile}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	stub := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	srv := New(cfg, logger, auth.NewAuthenticator(nil, nil, nil), stub)

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
