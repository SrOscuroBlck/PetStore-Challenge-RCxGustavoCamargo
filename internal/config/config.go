package config

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
)

type Config struct {
	HTTPAddr string
	LogLevel slog.Level
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

	return Config{
		HTTPAddr: httpAddr,
		LogLevel: logLevel,
	}, nil
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
