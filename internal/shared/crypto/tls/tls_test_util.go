// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"crypto/tls"
	http "net/http"

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
		Timeout: cryptoutilSharedMagic.CipherDefaultTimeout,
	}
}
