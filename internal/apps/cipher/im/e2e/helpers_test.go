// Copyright (c) 2025 Justin Cranford
//
// Shared test helpers for cipher-im e2e tests.

package e2e_test

import (
	"crypto/tls"
	"net/http"
	"time"
)

const (
	// testHTTPClientTimeout is the timeout for HTTP client requests.
	testHTTPClientTimeout = 10 * time.Second
)

// createHTTPSClient creates HTTP client with TLS verification disabled for testing.
func createHTTPSClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // Test environment only
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   testHTTPClientTimeout,
	}
}
