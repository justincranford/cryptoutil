// Copyright (c) 2025 Justin Cranford
//

// Package contract provides cross-service contract tests for the cryptoutil framework.
// Call RunContractTests from service integration tests to verify framework behavior consistency.
//
// Contract tests verify that ALL cryptoutil services implement consistent framework behavior:
// - Health endpoints return expected status codes and response bodies.
// - Server ports are dynamically allocated and properly isolated.
// - Response formats are consistent across all services.
//
// Usage in service integration tests:
//
//	func TestContractCompliance(t *testing.T) {
//	   t.Parallel()
//	   contract.RunContractTests(t, testServer)
//	}
package contract

import (
	"context"
	http "net/http"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
)

// ServiceServer is a type alias for the framework's ServiceServer contract.
type ServiceServer = cryptoutilAppsTemplateServiceServer.ServiceServer

// defaultContractRequestTimeout is the HTTP timeout for contract test requests.
const defaultContractRequestTimeout = 30 * time.Second

// RunContractTests runs all framework contracts against a live ServiceServer.
// Call this from each service's integration test suite after server startup.
// Verifies that ALL cryptoutil services implement consistent framework behavior.
//
// Note: RunReadyzNotReadyContract is excluded because it temporarily modifies
// server state; call it separately when safe (i.e., no concurrent readyz checks).
func RunContractTests(t *testing.T, server ServiceServer) {
	t.Helper()

	t.Run("health_contracts", func(t *testing.T) {
		t.Parallel()

		RunHealthContracts(t, server)
	})

	t.Run("server_contracts", func(t *testing.T) {
		t.Parallel()

		RunServerContracts(t, server)
	})

	t.Run("response_format_contracts", func(t *testing.T) {
		t.Parallel()

		RunResponseFormatContracts(t, server)
	})
}

// newTLSHTTPClient creates a TLS-skipping HTTP client for contract tests.
// Safe for test-only use against auto-generated self-signed certificates.
func newTLSHTTPClient(t *testing.T) *http.Client {
	t.Helper()

	tlsConfig, err := cryptoutilSharedCryptoTls.NewClientConfig(&cryptoutilSharedCryptoTls.ClientConfigOptions{
		SkipVerify: true, //nolint:gosec // test-only: auto-generated self-signed test certificates
	})
	if err != nil {
		t.Fatalf("contract: failed to create TLS client config: %v", err)
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   tlsConfig.TLSConfig,
			DisableKeepAlives: true, // Close connection after each request to prevent server shutdown hang
		},
		Timeout: defaultContractRequestTimeout,
	}
}

// newContractRequest creates an HTTP GET request with background context.
func newContractRequest(t *testing.T, url string) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("contract: failed to create request for %s: %v", url, err)
	}

	return req
}
