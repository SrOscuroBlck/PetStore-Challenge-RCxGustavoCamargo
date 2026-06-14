package config

import (
	"errors"
	"log/slog"
	"testing"
)

func TestLoad_MissingHTTPAddrIsRejectedAndNamed(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")

	_, err := Load()

	var missing *MissingConfigError
	if !errors.As(err, &missing) {
		t.Fatalf("expected *MissingConfigError, got %v", err)
	}
	if missing.Key != "HTTP_ADDR" {
		t.Fatalf("expected the error to name HTTP_ADDR, got %q", missing.Key)
	}
}

func TestLoad_ParsesAddrAndLevel(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":8443")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPAddr != ":8443" {
		t.Fatalf("expected HTTPAddr :8443, got %q", cfg.HTTPAddr)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Fatalf("expected LogLevel debug, got %v", cfg.LogLevel)
	}
}

func TestLoad_RejectsMalformedHTTPAddr(t *testing.T) {
	t.Setenv("HTTP_ADDR", "not-an-address")

	_, err := Load()

	var invalid *InvalidConfigError
	if !errors.As(err, &invalid) {
		t.Fatalf("expected *InvalidConfigError, got %v", err)
	}
	if invalid.Key != "HTTP_ADDR" {
		t.Fatalf("expected the error to name HTTP_ADDR, got %q", invalid.Key)
	}
}

func TestLoad_DefaultsLogLevelToInfo(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":8443")
	t.Setenv("LOG_LEVEL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Fatalf("expected default LogLevel info, got %v", cfg.LogLevel)
	}
}

func TestLoad_RejectsInvalidLogLevel(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":8443")
	t.Setenv("LOG_LEVEL", "verbose")

	_, err := Load()

	var invalid *InvalidConfigError
	if !errors.As(err, &invalid) {
		t.Fatalf("expected *InvalidConfigError, got %v", err)
	}
	if invalid.Key != "LOG_LEVEL" {
		t.Fatalf("expected the error to name LOG_LEVEL, got %q", invalid.Key)
	}
}
