// Copyright (c) 2025 Justin Cranford

//go:build e2e

package test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// CreateHTTPClient creates an HTTP client with standard test configuration.
func CreateHTTPClient() *http.Client {
	return &http.Client{Timeout: cryptoutilMagic.TestTimeoutHTTPClient}
}

// CreateInsecureHTTPClient creates an HTTP client that skips TLS verification (for self-signed certificates).
func CreateInsecureHTTPClient() *http.Client {
	return &http.Client{
		Timeout: cryptoutilMagic.DockerHTTPClientTimeoutSeconds * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // G402: E2E testing with self-signed certs
		},
	}
}

// CreateHTTPGetRequest creates an HTTP GET request with context.
func CreateHTTPGetRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP GET request: %w", err)
	}

	return req, nil
}
