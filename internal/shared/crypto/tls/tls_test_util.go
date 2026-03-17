// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	http "net/http"
	"os"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// NewClientForTest creates an HTTP client configured for testing with insecure TLS.
// Uses TLS 1.3 to match server requirements.
func NewClientForTest() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS13,
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: cryptoutilSharedMagic.IMDefaultTimeout,
	}
}

// NewClientForTestWithCA creates an HTTP client that verifies TLS using the provided CA cert file.
// Use this in E2E tests where a pki-init service has generated a shared root CA certificate.
func NewClientForTestWithCA(caCertPath string) *http.Client {
	caCert, err := os.ReadFile(caCertPath) //nolint:gosec // CA cert path is from trusted test config.
	if err != nil {
		panic(fmt.Sprintf("NewClientForTestWithCA: failed to read CA cert %q: %v", caCertPath, err))
	}

	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(caCert) {
		panic(fmt.Sprintf("NewClientForTestWithCA: no valid certs parsed from %q", caCertPath))
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    rootCAs,
			},
			DisableKeepAlives: true,
		},
		Timeout: cryptoutilSharedMagic.IMDefaultTimeout,
	}
}
