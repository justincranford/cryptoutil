// Copyright (c) 2025 Justin Cranford

// Package e2e_helpers provides reusable end-to-end testing helpers for all cryptoutil services.
// Extracted from sm-im implementation to support 9-service migration.
package e2e_helpers

import (
	"crypto/tls"
	"crypto/x509"
	http "net/http"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// CreateHTTPClient creates an HTTPS client using the provided certificate pool.
// Reusable for all services using TLS certificates in tests.
func CreateHTTPClient(t *testing.T, certPool *x509.CertPool) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    certPool,
			},
			DisableKeepAlives: true,
		},
		Timeout: cryptoutilSharedMagic.IMDefaultTimeout, // Increased for concurrent test execution.
	}
}

// CreateHTTPClientWithTimeout creates an HTTPS client with custom timeout.
// Useful for tests requiring longer timeouts (e.g., race detector mode).
func CreateHTTPClientWithTimeout(t *testing.T, timeout int64, certPool *x509.CertPool) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    certPool,
			},
			DisableKeepAlives: true,
		},
		Timeout: cryptoutilSharedMagic.IMDefaultTimeout * time.Duration(timeout),
	}
}
