// Copyright (c) 2025-2026 Justin Cranford.

// Package test_help_api provides HTTP client helpers, health checks, request assertions,
// and reusable HTTP mock servers for integration and E2E test suites.
// It handles health endpoint checking, HTTP request/response validation, and mock server setup.
//
// Consumed by:
//   - test_orch_e2e: health checks, HTTP clients
//   - test_orch_integration: health checks, HTTP clients, assertions
//   - API test suites: HTTP mocks and assertions
package test_help_api

import (
	"fmt"
	"io"
	http "net/http"

	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// HealthClient is a test-only HTTPS client for hitting service health endpoints.
type HealthClient struct {
	publicBaseURL string
	adminBaseURL  string
	client        *http.Client
}

// NewHealthClient creates a new HealthClient using TLS skip-verify (safe for auto-generated test certs).
func NewHealthClient(publicBaseURL, adminBaseURL string) *HealthClient {
	tlsConfig, err := cryptoutilSharedCryptoTls.NewClientConfig(&cryptoutilSharedCryptoTls.ClientConfigOptions{
		SkipVerify: true, //nolint:gosec // test-only: auto-generated self-signed test certificates
	})

	transport := &http.Transport{
		DisableKeepAlives: true, // Close connection after each request to prevent server shutdown hang
	}

	if err == nil {
		transport.TLSClientConfig = tlsConfig.TLSConfig
	}

	return &HealthClient{
		publicBaseURL: publicBaseURL,
		adminBaseURL:  adminBaseURL,
		client: &http.Client{
			Timeout:   cryptoutilSharedMagic.IMDefaultTimeout,
			Transport: transport,
		},
	}
}

// Livez calls the admin livez endpoint.
func (h *HealthClient) Livez() (*http.Response, error) {
	path := h.adminBaseURL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath

	resp, err := h.client.Get(path) //nolint:noctx // test helper: no context needed for health polling
	if err != nil {
		return nil, fmt.Errorf("livez request failed: %w", err)
	}

	return resp, nil
}

// Readyz calls the admin readyz endpoint.
func (h *HealthClient) Readyz() (*http.Response, error) {
	path := h.adminBaseURL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath

	resp, err := h.client.Get(path) //nolint:noctx // test helper: no context needed for health polling
	if err != nil {
		return nil, fmt.Errorf("readyz request failed: %w", err)
	}

	return resp, nil
}

// ServiceHealth calls the public service-path health endpoint.
func (h *HealthClient) ServiceHealth() (*http.Response, error) {
	path := h.publicBaseURL + cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath + "/health"

	resp, err := h.client.Get(path) //nolint:noctx // test helper: no context needed for health polling
	if err != nil {
		return nil, fmt.Errorf("servicehealth request failed: %w", err)
	}

	return resp, nil
}

// BrowserHealth calls the public browser-path health endpoint.
func (h *HealthClient) BrowserHealth() (*http.Response, error) {
	path := h.publicBaseURL + cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath + "/health"

	resp, err := h.client.Get(path) //nolint:noctx // test helper: no context needed for health polling
	if err != nil {
		return nil, fmt.Errorf("browserhealth request failed: %w", err)
	}

	return resp, nil
}

// PublicHealth calls the public browser-path health endpoint (alias for BrowserHealth).
func (h *HealthClient) PublicHealth() (*http.Response, error) {
	return h.BrowserHealth()
}

// DrainAndClose reads and discards all bytes from resp.Body then closes it.
func (h *HealthClient) DrainAndClose(resp *http.Response) error {
	if resp == nil || resp.Body == nil {
		return nil
	}

	_, _ = io.ReadAll(resp.Body)

	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("close response body: %w", err)
	}

	return nil
}
