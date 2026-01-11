// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	cryptoutilMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestE2E_HealthChecks validates /health endpoint for all instances (external clients use public endpoint).
func TestE2E_HealthChecks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		publicURL string
	}{
		{sqliteContainer, sqlitePublicURL},
		{postgres1Container, postgres1PublicURL},
		{postgres2Container, postgres2PublicURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Public health check (external clients MUST use this endpoint).
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			healthURL := tt.publicURL + cryptoutilMagic.CipherE2EHealthEndpoint
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
			require.NoError(t, err, "Creating health check request should succeed")

			healthResp, err := sharedHTTPClient.Do(req)
			require.NoError(t, err, "Health check should succeed for %s", tt.name)
			require.NoError(t, healthResp.Body.Close())
			require.Equal(t, http.StatusOK, healthResp.StatusCode,
				"%s should return 200 OK for /health", tt.name)
		})
	}
}

// TestE2E_TelemetryServices validates otel-collector and Grafana LGTM containers are healthy.
func TestE2E_TelemetryServices(t *testing.T) {
	t.Parallel()

	t.Run(cryptoutilMagic.CipherE2EOtelCollectorContainer, func(t *testing.T) {
		t.Parallel()

		// OpenTelemetry Collector health check via HTTP endpoint.
		// The otel-collector-contrib image exposes a health check extension on port 13133.
		// However, this port is not exposed to the host in the compose.yml for security.
		// We verify the container is running and accepting OTLP connections by checking
		// that cipher-im services are successfully sending telemetry (no connection errors).
		// For E2E, we rely on docker compose health checks and the fact that services start.
		// A more robust check would use `docker exec` to query the internal health endpoint.

		// Verify the service is accessible by attempting a connection (will fail, but proves routing).
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		client := &http.Client{Timeout: 2 * time.Second}
		// Note: OTLP gRPC port 4317 won't respond to HTTP, but connection attempt proves DNS resolution.
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, otelCollectorURL, nil)
		require.NoError(t, err, "Creating OTLP request should succeed")

		resp, err := client.Do(req)
		if resp != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
		// We expect an error (gRPC port doesn't speak HTTP), but NO "connection refused" or "no such host".
		// This proves the container is running and network routing works.
		require.Error(t, err, "OTLP gRPC port should not respond to HTTP GET")
		require.NotContains(t, err.Error(), "connection refused", "Container should be running")
		require.NotContains(t, err.Error(), "no such host", "Container DNS should resolve")
	})

	t.Run(cryptoutilMagic.CipherE2EGrafanaContainer, func(t *testing.T) {
		t.Parallel()

		// Grafana HTTP API health check.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, grafanaURL+"/api/health", nil)
		require.NoError(t, err, "Creating Grafana health request should succeed")

		resp, err := client.Do(req)
		require.NoError(t, err, "Grafana health endpoint should respond")
		require.NoError(t, resp.Body.Close())
		require.Equal(t, http.StatusOK, resp.StatusCode, "Grafana should return 200 OK")
	})
}

// TestE2E_CrossInstanceIsolation verifies each instance has independent tenant/user state.
// Creates a user in one instance and verifies it's NOT visible in other instances.
func TestE2E_CrossInstanceIsolation(t *testing.T) {
	t.Parallel()

	instances := []struct {
		name      string
		publicURL string
	}{
		{sqliteContainer, sqlitePublicURL},
		{postgres1Container, postgres1PublicURL},
		{postgres2Container, postgres2PublicURL},
	}

	for _, inst := range instances {
		t.Run(inst.name, func(t *testing.T) {
			t.Parallel()

			// Create a unique user in this instance.
			username := fmt.Sprintf("user_%s_%d", inst.name, time.Now().UnixNano())
			password := "TestPassword123!"

			// Register user in current instance.
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			registerURL := inst.publicURL + "/service/api/v1/users/register"
			registerBody := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, registerURL, bytes.NewBufferString(registerBody))
			require.NoError(t, err, "Creating registration request should succeed")
			req.Header.Set("Content-Type", "application/json")

			resp, err := sharedHTTPClient.Do(req)
			require.NoError(t, err, "User registration should succeed in %s", inst.name)
			require.NoError(t, resp.Body.Close())
			require.Equal(t, http.StatusCreated, resp.StatusCode,
				"User should be created in %s", inst.name)

			// Verify user can login in the SAME instance.
			loginURL := inst.publicURL + "/service/api/v1/users/login"
			loginBody := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)

			loginReq, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewBufferString(loginBody))
			require.NoError(t, err, "Creating login request should succeed")
			loginReq.Header.Set("Content-Type", "application/json")

			loginResp, err := sharedHTTPClient.Do(loginReq)
			require.NoError(t, err, "Login should succeed in same instance %s", inst.name)
			require.NoError(t, loginResp.Body.Close())
			require.Equal(t, http.StatusOK, loginResp.StatusCode,
				"User should login successfully in %s", inst.name)

			// Verify user does NOT exist in OTHER instances (tenant isolation).
			for _, otherInst := range instances {
				if otherInst.name == inst.name {
					continue // Skip same instance
				}

				otherLoginURL := otherInst.publicURL + "/service/api/v1/users/login"
				otherLoginReq, err := http.NewRequestWithContext(ctx, http.MethodPost, otherLoginURL, bytes.NewBufferString(loginBody))
				require.NoError(t, err, "Creating other login request should succeed")
				otherLoginReq.Header.Set("Content-Type", "application/json")

				otherLoginResp, err := sharedHTTPClient.Do(otherLoginReq)
				require.NoError(t, err, "Login attempt should complete in %s", otherInst.name)
				require.NoError(t, otherLoginResp.Body.Close())
				require.NotEqual(t, http.StatusOK, otherLoginResp.StatusCode,
					"User from %s should NOT exist in %s (tenant isolation)",
					inst.name, otherInst.name)
			}
		})
	}
}
