// Copyright (c) 2025 Justin Cranford
//

package contract

import (
	json "encoding/json"
	"io"
	http "net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilTestingHealthclient "cryptoutil/internal/apps/template/service/testing/healthclient"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// RunHealthContracts verifies all health endpoint contracts for the given service.
// Tests 8 contracts:
//  1. livez returns HTTP 200
//  2. livez response body has status=alive
//  3. readyz (ready) returns HTTP 200
//  4. readyz (ready) response body has status=ready
//  5. browser health returns HTTP 200
//  6. browser health response body has status=healthy
//  7. service health returns HTTP 200
//  8. service health response body has status=healthy
func RunHealthContracts(t *testing.T, server ServiceServer) {
	t.Helper()

	hc := cryptoutilTestingHealthclient.NewHealthClient(server.PublicBaseURL(), server.AdminBaseURL())

	tests := []struct {
		name       string
		fetch      func() (*http.Response, error)
		wantCode   int
		wantStatus string
	}{
		{
			name:       "livez_returns_200_with_status_alive",
			fetch:      hc.Livez,
			wantCode:   http.StatusOK,
			wantStatus: "alive",
		},
		{
			name:       "readyz_ready_returns_200_with_status_ready",
			fetch:      hc.Readyz,
			wantCode:   http.StatusOK,
			wantStatus: "ready",
		},
		{
			name:       "browser_health_returns_200_with_status_healthy",
			fetch:      hc.BrowserHealth,
			wantCode:   http.StatusOK,
			wantStatus: cryptoutilSharedMagic.DockerServiceHealthHealthy,
		},
		{
			name:       "service_health_returns_200_with_status_healthy",
			fetch:      hc.ServiceHealth,
			wantCode:   http.StatusOK,
			wantStatus: cryptoutilSharedMagic.DockerServiceHealthHealthy,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp, err := tc.fetch()
			require.NoError(t, err, "health endpoint request should succeed")

			defer func() { require.NoError(t, resp.Body.Close()) }()

			assert.Equal(t, tc.wantCode, resp.StatusCode, "unexpected HTTP status code")

			body, readErr := io.ReadAll(resp.Body)
			require.NoError(t, readErr, "reading response body should not fail")

			var result map[string]string

			require.NoError(t, json.Unmarshal(body, &result), "response body must be valid JSON")
			assert.Equal(t, tc.wantStatus, result[cryptoutilSharedMagic.StringStatus], "response body status field mismatch")
		})
	}
}

// RunReadyzNotReadyContract verifies that readyz returns 503 when the server is not ready.
// This contract tests that the framework correctly rejects traffic from a not-ready server.
//
// WARNING: This function temporarily sets server.SetReady(false) then restores it.
// Do NOT call this in parallel with other tests that depend on readyz returning 200.
// Always call this as a non-parallel test function.
func RunReadyzNotReadyContract(t *testing.T, server ServiceServer) {
	t.Helper()

	server.SetReady(false)
	defer server.SetReady(true)

	hc := cryptoutilTestingHealthclient.NewHealthClient(server.PublicBaseURL(), server.AdminBaseURL())

	resp, err := hc.Readyz()
	require.NoError(t, err, "readyz request should not fail even when server is not ready")

	defer func() { require.NoError(t, resp.Body.Close()) }()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode, "readyz should return 503 when server is not ready")
}
