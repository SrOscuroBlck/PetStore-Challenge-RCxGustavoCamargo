package tlscert

import (
	"crypto/tls"
	"crypto/x509"
	"testing"
)

func TestGenerate_ProducesUsableKeyPair(t *testing.T) {
	certPEM, keyPEM, err := Generate([]string{"localhost", "127.0.0.1"})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("the cert and key must form a usable TLS pair: %v", err)
	}

	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	if err := leaf.VerifyHostname("localhost"); err != nil {
		t.Fatalf("certificate must be valid for localhost: %v", err)
	}
	if len(leaf.IPAddresses) == 0 {
		t.Fatal("certificate must carry the 127.0.0.1 SAN")
	}
}
