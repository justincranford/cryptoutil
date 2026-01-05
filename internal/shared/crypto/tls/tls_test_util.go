// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"crypto/tls"
	"net/http"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// NewClientForTest creates an HTTP client configured for testing with insecure TLS.
func NewClientForTest() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: cryptoutilMagic.CipherDefaultTimeout,
	}
}