// Copyright (c) 2025 Justin Cranford
//
// This file implements Docker Compose integration tests for cipher-im.
// Uses shared TestMain lifecycle from testmain_e2e_test.go (stack already running).
//
// Per 03-02.testing.instructions.md:
// - Table-driven tests with t.Parallel() for orthogonal scenarios
// - TestMain pattern for heavyweight service startup (in testmain_e2e_test.go)
// - Coverage targets: ≥98% for infrastructure code

package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

const (
	httpClientTimeout = 10 * time.Second
)

// TestDockerComposeHealthEndpoints validates health endpoints for all running instances.
// Stack is already running via TestMain in testmain_e2e_test.go.
func TestDockerComposeHealthEndpoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
	}{
		{name: "cipher-im-sqlite livez", url: fmt.Sprintf("https://127.0.0.1:%d/admin/v1/livez", cryptoutilMagic.CipherE2ESQLiteAdminPort)},
		{name: "cipher-im-pg-1 livez", url: fmt.Sprintf("https://127.0.0.1:%d/admin/v1/livez", cryptoutilMagic.CipherE2EPostgreSQL1AdminPort)},
		{name: "cipher-im-pg-2 livez", url: fmt.Sprintf("https://127.0.0.1:%d/admin/v1/livez", cryptoutilMagic.CipherE2EPostgreSQL2AdminPort)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, tt.url, nil)
			require.NoError(t, err, "Creating request should succeed")

			resp, err := sharedHTTPClient.Do(req)
			require.NoError(t, err, "GET %s should succeed", tt.url)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, http.StatusOK, resp.StatusCode, "should return 200 OK")

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err, "reading response body should succeed")

			var healthStatus map[string]string

			err = json.Unmarshal(body, &healthStatus)
			require.NoError(t, err, "unmarshaling JSON should succeed")

			require.Equal(t, "alive", healthStatus["status"], "status should be 'alive'")
		})
	}
}

// TestAllInstancesLivezAndReadyz validates livez and readyz endpoints for all instances.
// Stack is already running via TestMain in testmain_e2e_test.go.
func TestAllInstancesLivezAndReadyz(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		adminPort     int
		containerName string
	}{
		{name: "cipher-im-sqlite", adminPort: cryptoutilMagic.CipherE2ESQLiteAdminPort, containerName: sqliteContainer},
		{name: "cipher-im-pg-1", adminPort: cryptoutilMagic.CipherE2EPostgreSQL1AdminPort, containerName: postgres1Container},
		{name: "cipher-im-pg-2", adminPort: cryptoutilMagic.CipherE2EPostgreSQL2AdminPort, containerName: postgres2Container},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
			defer cancel()

			// Validate livez endpoint.
			livezURL := fmt.Sprintf("https://%s:%d/admin/v1/livez", cryptoutilMagic.IPv4Loopback, tt.adminPort)
			livezReq, err := http.NewRequestWithContext(ctx, http.MethodGet, livezURL, nil)
			require.NoError(t, err, "Creating livez request should succeed")

			resp, err := sharedHTTPClient.Do(livezReq)
			require.NoError(t, err, "GET livez should succeed")

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, http.StatusOK, resp.StatusCode, "livez should return 200 OK")

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err, "reading livez body should succeed")

			var status map[string]string

			err = json.Unmarshal(body, &status)
			require.NoError(t, err, "unmarshaling livez JSON should succeed")

			require.Equal(t, "alive", status["status"], "livez status should be 'alive'")

			// Validate readyz endpoint.
			readyzURL := fmt.Sprintf("https://%s:%d/admin/v1/readyz", cryptoutilMagic.IPv4Loopback, tt.adminPort)
			readyzReq, err := http.NewRequestWithContext(ctx, http.MethodGet, readyzURL, nil)
			require.NoError(t, err, "Creating readyz request should succeed")

			resp, err = sharedHTTPClient.Do(readyzReq)
			require.NoError(t, err, "GET readyz should succeed")

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, http.StatusOK, resp.StatusCode, "readyz should return 200 OK")

			body, err = io.ReadAll(resp.Body)
			require.NoError(t, err, "reading readyz body should succeed")

			err = json.Unmarshal(body, &status)
			require.NoError(t, err, "unmarshaling readyz JSON should succeed")

			require.Equal(t, "ready", status["status"], "readyz status should be 'ready'")
		})
	}
}
