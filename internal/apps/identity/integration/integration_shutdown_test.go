// Copyright (c) 2025 Justin Cranford
//
//

//go:build integration

// Package integration provides integration tests for the identity server.
// These tests require real network listeners and use hardcoded ports,
// so they must be run separately from unit tests using:
//
//	go test -tags=integration ./internal/identity/integration/...
//
// Do NOT run these with `go test ./...` as they will conflict with parallel tests.
package integration

import (
	"context"
	json "encoding/json"
	"io"
	http "net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	testify "github.com/stretchr/testify/require"
)

func TestUnauthorizedAccess(t *testing.T) {
	t.Parallel()
	// Sequential execution to avoid port conflicts with hardcoded ports.
	testMutex.Lock()

	servers, cancel := setupTestServers(t)

	defer func() {
		shutdownTestServers(t, servers)
		cancel()
		testMutex.Unlock()
	}()

	endpoints := []string{
		"/api/v1/protected/resource",
		"/api/v1/admin/users",
		"/api/v1/admin/metrics",
	}

	for _, endpoint := range endpoints {
		t.Run("Unauthorized_"+endpoint, func(t *testing.T) {
			resourceURL := testRSBaseURL + endpoint
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, resourceURL, nil)
			testify.NoError(t, err, "Failed to create resource request")

			// Don't set Authorization header.
			resp, err := servers.httpClient.Do(req)
			testify.NoError(t, err, "Resource request failed")

			defer func() {
				defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test code cleanup
			}()

			testify.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should return 401 Unauthorized")
		})
	}
}

// TestGracefulShutdown verifies servers shut down cleanly.
//
// Validates requirements:
// - R03-03: Integration tests clean up resources.
func TestGracefulShutdown(t *testing.T) {
	t.Parallel()
	// Sequential execution to avoid port conflicts with hardcoded ports.
	testMutex.Lock()
	defer testMutex.Unlock()

	servers, cancel := setupTestServers(t)

	// Start servers normally.
	time.Sleep(serverStartDelay)

	// Verify servers are running.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, testAuthZBaseURL+"/health", nil)
	testify.NoError(t, err, "Failed to create health check request")

	resp, err := servers.httpClient.Do(req)
	testify.NoError(t, err, "Health check failed before shutdown")

	if resp != nil && resp.Body != nil {
		defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test code cleanup
	}

	testify.Equal(t, http.StatusOK, resp.StatusCode, "Server should be healthy before shutdown")

	// Cancel context to trigger shutdown.
	cancel()

	// Shutdown servers gracefully.
	shutdownTestServers(t, servers)

	// Verify servers stopped (connection should fail).
	time.Sleep(100 * time.Millisecond)

	req2, err := http.NewRequestWithContext(context.Background(), http.MethodGet, testAuthZBaseURL+"/health", nil)
	testify.NoError(t, err, "Failed to create health check request")

	resp2, err := servers.httpClient.Do(req2)
	if resp2 != nil && resp2.Body != nil {
		defer func() { _ = resp2.Body.Close() }() //nolint:errcheck // Test code cleanup
	}

	testify.Error(t, err, "Connection should fail after shutdown")
}

// getTestAccessToken helper function to obtain an access token for testing.
func getTestAccessToken(t *testing.T, servers *testServers, scope string) string {
	t.Helper()

	// Simplified token generation for testing - in real integration tests,
	// this would go through the full OAuth flow.
	tokenURL := testAuthZBaseURL + "/oauth2/v1/token"
	tokenData := url.Values{}
	tokenData.Set("grant_type", "client_credentials")
	tokenData.Set("client_id", testClientID)
	tokenData.Set("client_secret", testClientSecret)
	tokenData.Set("scope", scope)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, tokenURL, strings.NewReader(tokenData.Encode()))
	testify.NoError(t, err, "Failed to create token request")

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := servers.httpClient.Do(req)
	testify.NoError(t, err, "Token request failed")

	defer func() {
		defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test code cleanup
	}()

	testify.Equal(t, http.StatusOK, resp.StatusCode, "Token request should succeed")

	body, err := io.ReadAll(resp.Body)
	testify.NoError(t, err, "Failed to read token response")

	var tokenResponse map[string]any

	err = json.Unmarshal(body, &tokenResponse)
	testify.NoError(t, err, "Failed to decode token response")

	accessToken, ok := tokenResponse["access_token"].(string)
	testify.True(t, ok, "Access token should be present")
	testify.NotEmpty(t, accessToken, "Access token should not be empty")

	return accessToken
}
