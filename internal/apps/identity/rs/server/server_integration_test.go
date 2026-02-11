// Copyright (c) 2025 Justin Cranford

package server

import (
	"context"
	"crypto/tls"
	"io"
	http "net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRSServer_Lifecycle(t *testing.T) {
	t.Parallel()
	requireTestSetup(t)

	// Server should be running and ready from TestMain.
	require.NotNil(t, testServer)
	require.Greater(t, testServer.PublicPort(), 0)
	require.Greater(t, testServer.AdminPort(), 0)
}

func TestRSServer_PortAllocation(t *testing.T) {
	t.Parallel()
	requireTestSetup(t)

	// Dynamic ports should be allocated (not hardcoded).
	publicPort := testServer.PublicPort()
	adminPort := testServer.AdminPort()

	require.Greater(t, publicPort, 0, "Public port should be allocated")
	require.Greater(t, adminPort, 0, "Admin port should be allocated")
	require.NotEqual(t, publicPort, adminPort, "Ports should be different")
}

func TestRSServer_TemplateServices(t *testing.T) {
	t.Parallel()
	requireTestSetup(t)

	// Verify template services are initialized.
	require.NotNil(t, testServer.DB(), "Database should be initialized")
	require.NotNil(t, testServer.JWKGen(), "JWK generation service should be initialized")
	require.NotNil(t, testServer.Telemetry(), "Telemetry service should be initialized")
	require.NotNil(t, testServer.Barrier(), "Barrier service should be initialized")
	require.NotNil(t, testServer.App(), "Application should be initialized")
}

func TestRSServer_PublicHealth(t *testing.T) {
	t.Parallel()
	requireTestSetup(t)

	client := &http.Client{
		Timeout: httpTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment uses self-signed certs.
			},
		},
	}

	tests := []struct {
		name     string
		path     string
		wantCode int
	}{
		{"health", "/health", http.StatusOK},
		{"livez", "/livez", http.StatusOK},
		{"readyz", "/readyz", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := testBaseURL + tt.path

			ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)

			defer func() {
				_ = resp.Body.Close()
			}()

			require.Equal(t, tt.wantCode, resp.StatusCode)

			// Read body to verify response.
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.NotEmpty(t, body)
		})
	}
}

func TestRSServer_AccessorMethods(t *testing.T) {
	t.Parallel()
	requireTestSetup(t)

	// Test accessor methods.
	require.NotNil(t, testServer.Config())
	require.NotEmpty(t, testServer.PublicBaseURL())
	require.NotEmpty(t, testServer.AdminBaseURL())
}

// HTTP timeout for test requests.
const httpTimeout = 5 * time.Second
