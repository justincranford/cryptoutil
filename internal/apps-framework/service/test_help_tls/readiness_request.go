// Copyright (c) 2025-2026 Justin Cranford.
package test_help_tls

import (
	"fmt"
	http "net/http"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// DoReadinessRequest waits through transient fail-closed 503 observations using
// the caller's existing TLS client and deadline. Other statuses and transport
// errors are returned immediately for the test to validate. The caller owns
// the successful response body; all retried response bodies are closed.
func DoReadinessRequest(client *http.Client, request *http.Request) (*http.Response, error) {
	if client == nil || request == nil {
		return nil, fmt.Errorf("readiness client and request must not be nil")
	}

	if request.Method != http.MethodGet {
		return nil, fmt.Errorf("readiness request must use GET")
	}

	if request.URL == nil || request.URL.Scheme != cryptoutilSharedMagic.ProtocolHTTPS || request.URL.User != nil {
		return nil, fmt.Errorf("readiness request requires a loopback HTTPS endpoint without userinfo")
	}

	switch request.URL.Hostname() {
	case cryptoutilSharedMagic.IPv4Loopback, cryptoutilSharedMagic.IPv6Loopback, cryptoutilSharedMagic.HostnameLocalhost:
	default:
		return nil, fmt.Errorf("readiness request requires an explicitly allowed loopback host")
	}

	if err := request.Context().Err(); err != nil {
		return nil, fmt.Errorf("readiness request context: %w", err)
	}

	_, ok := request.Context().Deadline()
	if !ok {
		return nil, fmt.Errorf("readiness request requires a bounded deadline")
	}
	// Do not mutate the shared fixture client or follow redirects off loopback.
	localClient := *client
	localClient.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}

	ticker := time.NewTicker(cryptoutilSharedMagic.TestServerWaitTickerInterval)
	defer ticker.Stop()

	for {
		// #nosec G704 -- explicit HTTPS loopback allowlist above; redirects prohibited; verified fixture TLS and deadline retained.
		response, err := localClient.Do(request.Clone(request.Context()))
		if err != nil {
			return nil, fmt.Errorf("request readiness: %w", err)
		}

		if response.StatusCode != http.StatusServiceUnavailable {
			return response, nil
		}

		if err := response.Body.Close(); err != nil {
			return nil, fmt.Errorf("close transient readiness response: %w", err)
		}

		select {
		case <-request.Context().Done():
			return nil, fmt.Errorf("wait for readiness response: %w", request.Context().Err())
		case <-ticker.C:
		}
	}
}
