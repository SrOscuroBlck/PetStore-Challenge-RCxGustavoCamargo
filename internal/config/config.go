package config

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	"roboticCrewChallenge/internal/platform/crypto"
)

type Config struct {
	HTTPAddr         string
	LogLevel         slog.Level
	DatabaseURL      string
	PIIEncryptionKey []byte
}

type MissingConfigError struct {
	Key string
}

func (e *MissingConfigError) Error() string {
	return fmt.Sprintf("required configuration %q is not set", e.Key)
}

type InvalidConfigError struct {
	Key   string
	Value string
	Hint  string
}

func (e *InvalidConfigError) Error() string {
	return fmt.Sprintf("configuration %q has invalid value %q: %s", e.Key, e.Value, e.Hint)
}

func Load() (Config, error) {
	httpAddr, err := requireEnv("HTTP_ADDR")
	if err != nil {
		return Config{}, err
	}
	if _, _, err := net.SplitHostPort(httpAddr); err != nil {
		return Config{}, &InvalidConfigError{
			Key:   "HTTP_ADDR",
			Value: httpAddr,
			Hint:  "expected host:port (e.g. :8443)",
		}
	}

	logLevel, err := parseLogLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		return Config{}, err
	}

	databaseURL, err := requireEnv("DATABASE_URL")
	if err != nil {
		return Config{}, err
	}

	piiKey, err := loadEncryptionKey("PII_ENCRYPTION_KEY")
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTPAddr:         httpAddr,
		LogLevel:         logLevel,
		DatabaseURL:      databaseURL,
		PIIEncryptionKey: piiKey,
	}, nil
}

func loadEncryptionKey(key string) ([]byte, error) {
	encoded, err := requireEnv(key)
	if err != nil {
		return nil, err
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, &InvalidConfigError{
			Key:   key,
			Value: "<redacted>",
			Hint:  "expected base64-encoded 32-byte key",
		}
	}
	if len(decoded) != crypto.KeySize {
		return nil, &InvalidConfigError{
			Key:   key,
			Value: "<redacted>",
			Hint:  fmt.Sprintf("expected %d bytes after base64 decode, got %d", crypto.KeySize, len(decoded)),
		}
	}
	return decoded, nil
}

func requireEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return "", &MissingConfigError{Key: key}
	}
	return value, nil
}

func parseLogLevel(value string) (slog.Level, error) {
	switch strings.ToLower(value) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, &InvalidConfigError{
			Key:   "LOG_LEVEL",
			Value: value,
			Hint:  "expected one of: debug, info, warn, error",
		}
	}
}
