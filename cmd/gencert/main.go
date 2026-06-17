// Command gencert writes a self-signed TLS certificate and key for local
// development into ./certs. It is a dev convenience only; production
// certificates come from the deployment's secrets.
package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"roboticCrewChallenge/internal/platform/tlscert"
)

const outputDir = "certs"

func main() {
	if err := run(); err != nil {
		slog.Error("gencert failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return err
	}

	certPEM, keyPEM, err := tlscert.Generate([]string{"localhost", "127.0.0.1", "::1"})
	if err != nil {
		return err
	}

	certPath := filepath.Join(outputDir, "cert.pem")
	keyPath := filepath.Join(outputDir, "key.pem")
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		return err
	}

	slog.Info("wrote self-signed TLS certificate", "cert", certPath, "key", keyPath)
	return nil
}
