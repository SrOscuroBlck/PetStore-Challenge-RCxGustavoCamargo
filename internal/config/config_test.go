package config

import (
	"encoding/base64"
	"errors"
	"log/slog"
	"testing"

	"roboticCrewChallenge/internal/platform/crypto"
)

func validKey() string {
	return base64.StdEncoding.EncodeToString(make([]byte, crypto.KeySize))
}

func setRequired(t *testing.T) {
	t.Helper()
	t.Setenv("HTTP_ADDR", ":8443")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/petstore")
	t.Setenv("PII_ENCRYPTION_KEY", validKey())
}

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
	setRequired(t)
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
	if len(cfg.PIIEncryptionKey) != crypto.KeySize {
		t.Fatalf("expected a %d-byte key, got %d", crypto.KeySize, len(cfg.PIIEncryptionKey))
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
	setRequired(t)
	t.Setenv("LOG_LEVEL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Fatalf("expected default LogLevel info, got %v", cfg.LogLevel)
	}
}

func TestLoad_MissingDatabaseURLIsRejectedAndNamed(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":8443")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("PII_ENCRYPTION_KEY", validKey())

	_, err := Load()

	var missing *MissingConfigError
	if !errors.As(err, &missing) {
		t.Fatalf("expected *MissingConfigError, got %v", err)
	}
	if missing.Key != "DATABASE_URL" {
		t.Fatalf("expected the error to name DATABASE_URL, got %q", missing.Key)
	}
}

func TestLoad_RejectsMalformedEncryptionKey(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":8443")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/petstore")
	t.Setenv("PII_ENCRYPTION_KEY", "not-valid-base64!!")

	_, err := Load()

	var invalid *InvalidConfigError
	if !errors.As(err, &invalid) {
		t.Fatalf("expected *InvalidConfigError, got %v", err)
	}
	if invalid.Key != "PII_ENCRYPTION_KEY" {
		t.Fatalf("expected the error to name PII_ENCRYPTION_KEY, got %q", invalid.Key)
	}
	if invalid.Value != "<redacted>" {
		t.Fatalf("expected the key value to be redacted, got %q", invalid.Value)
	}
}

func TestLoad_RejectsWrongLengthEncryptionKey(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":8443")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/petstore")
	t.Setenv("PII_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(make([]byte, 16)))

	_, err := Load()

	var invalid *InvalidConfigError
	if !errors.As(err, &invalid) {
		t.Fatalf("expected *InvalidConfigError, got %v", err)
	}
	if invalid.Key != "PII_ENCRYPTION_KEY" {
		t.Fatalf("expected the error to name PII_ENCRYPTION_KEY, got %q", invalid.Key)
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
