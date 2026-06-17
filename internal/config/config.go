package config

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"

	"roboticCrewChallenge/internal/platform/crypto"
)

type Config struct {
	HTTPAddr         string
	LogLevel         slog.Level
	DatabaseURL      string
	PIIEncryptionKey []byte
	RedisAddr        string
	MinIOEndpoint    string
	MinIOAccessKey   string
	MinIOSecretKey   string
	MinIOBucket      string
	MinIOUseSSL      bool
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

	redisAddr, err := requireEnv("REDIS_ADDR")
	if err != nil {
		return Config{}, err
	}

	minioEndpoint, err := requireEnv("MINIO_ENDPOINT")
	if err != nil {
		return Config{}, err
	}
	if strings.Contains(minioEndpoint, "://") {
		return Config{}, &InvalidConfigError{
			Key:   "MINIO_ENDPOINT",
			Value: minioEndpoint,
			Hint:  "expected host:port without a scheme; control TLS via MINIO_USE_SSL",
		}
	}
	minioAccessKey, err := requireEnv("MINIO_ACCESS_KEY")
	if err != nil {
		return Config{}, err
	}
	minioSecretKey, err := requireEnv("MINIO_SECRET_KEY")
	if err != nil {
		return Config{}, err
	}
	minioBucket, err := requireEnv("MINIO_BUCKET")
	if err != nil {
		return Config{}, err
	}
	minioUseSSL, err := parseBool("MINIO_USE_SSL", false)
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTPAddr:         httpAddr,
		LogLevel:         logLevel,
		DatabaseURL:      databaseURL,
		PIIEncryptionKey: piiKey,
		RedisAddr:        redisAddr,
		MinIOEndpoint:    minioEndpoint,
		MinIOAccessKey:   minioAccessKey,
		MinIOSecretKey:   minioSecretKey,
		MinIOBucket:      minioBucket,
		MinIOUseSSL:      minioUseSSL,
	}, nil
}

func parseBool(key string, fallback bool) (bool, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, &InvalidConfigError{
			Key:   key,
			Value: raw,
			Hint:  "expected a boolean (true, false, 1, 0)",
		}
	}
	return value, nil
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
