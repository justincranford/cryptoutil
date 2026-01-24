// Copyright (c) 2025 Justin Cranford

// Package e2e provides reusable end-to-end testing helpers for all cryptoutil services.
// Extracted from cipher-im implementation to support 9-service migration.
package e2e

import (
	"crypto/tls"
	http "net/http"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// CreateInsecureHTTPClient creates an HTTP client that trusts self-signed certificates.
// Reusable for all services using self-signed TLS certificates in tests.
//
// WARNING: Only use in test environments. Production code MUST validate certificates.
func CreateInsecureHTTPClient(t *testing.T) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: cryptoutilSharedMagic.CipherDefaultTimeout, // Increased for concurrent test execution.
	}
}

// CreateInsecureHTTPClientWithTimeout creates HTTP client with custom timeout.
// Useful for tests requiring longer timeouts (e.g., race detector mode).
func CreateInsecureHTTPClientWithTimeout(t *testing.T, timeout int64) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: cryptoutilSharedMagic.CipherDefaultTimeout * time.Duration(timeout),
	}
}
