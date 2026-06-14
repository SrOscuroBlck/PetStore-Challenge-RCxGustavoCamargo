package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"roboticCrewChallenge/internal/config"
)

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
	cfg := config.Config{HTTPAddr: "127.0.0.1:0", LogLevel: slog.LevelInfo}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := New(cfg, logger)

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
