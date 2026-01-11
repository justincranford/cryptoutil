// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestE2E_HealthChecks validates /admin/v1/livez and /admin/v1/readyz for all instances.
func TestE2E_HealthChecks(t *testing.T) {
	tests := []struct {
		name     string
		adminURL string
	}{
		{"SQLite Instance", sqliteAdminURL},
		{"PostgreSQL-1 Instance", postgres1AdminURL},
		{"PostgreSQL-2 Instance", postgres2AdminURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Livez (lightweight check - process alive?).
			livezResp, err := sharedHTTPClient.Get(tt.adminURL + "/admin/v1/livez")
			require.NoError(t, err)
			require.NoError(t, livezResp.Body.Close())
			require.Equal(t, http.StatusOK, livezResp.StatusCode)

			// Readyz (heavyweight check - dependencies healthy?).
			readyzResp, err := sharedHTTPClient.Get(tt.adminURL + "/admin/v1/readyz")
			require.NoError(t, err)
			require.NoError(t, readyzResp.Body.Close())
			require.Equal(t, http.StatusOK, readyzResp.StatusCode)
		})
	}
}

// TestE2E_TelemetryServices validates otel-collector and Grafana LGTM.
func TestE2E_TelemetryServices(t *testing.T) {
	t.Run("OtelCollector", func(t *testing.T) {
		// OTLP receiver health check (gRPC port 4317 is active).
		// TODO: Implement actual gRPC health check when otel-collector exposes it.
		// For now, we rely on docker compose health checks.
		t.Skip("OTLP gRPC health check requires grpc-health-probe tool")
	})

	t.Run("GrafanaLGTM", func(t *testing.T) {
		// Grafana HTTP API health check.
		client := &http.Client{Timeout: 5 * 5} // Standard HTTP client for Grafana.
		resp, err := client.Get(grafanaURL + "/api/health")
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestE2E_CrossInstanceIsolation verifies each instance has independent state.
func TestE2E_CrossInstanceIsolation(t *testing.T) {
	instances := []struct {
		name      string
		publicURL string
	}{
		{"SQLite", sqlitePublicURL},
		{"PostgreSQL-1", postgres1PublicURL},
		{"PostgreSQL-2", postgres2PublicURL},
	}

	for _, inst := range instances {
		t.Run(inst.name, func(t *testing.T) {
			t.Parallel()

			// Each instance should have independent tenant/user state.
			// TODO: Create tenant in instance A, verify NOT visible in instance B/C.
			// This requires implementing tenant CRUD operations first.
			t.Skip(fmt.Sprintf("Tenant isolation test for %s requires CRUD implementation", inst.name))
		})
	}
}
