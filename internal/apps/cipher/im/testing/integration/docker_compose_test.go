// Copyright (c) 2025 Justin Cranford
//
// This file implements Docker Compose full-stack lifecycle tests for cipher-im.
//
// Test Coverage:
// - Full stack deployment (3 cipher-im instances + PostgreSQL + Grafana OTEL + collector)
// - Health endpoint validation (livez, readyz)
// - Container lifecycle management (up, down, cleanup)
//
// Per 03-02.testing.instructions.md:
// - Table-driven tests with t.Parallel() for orthogonal scenarios
// - TestMain pattern for heavyweight service startup
// - Dynamic port allocation (port 0) for test isolation
// - Coverage targets: ≥98% for infrastructure code

package integration_test

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

const (
	dockerComposeFile = "../../../../../../cmd/cipher-im/docker-compose.yml"
	healthTimeout     = 90 * time.Second
	httpClientTimeout = 10 * time.Second
)

// TestMain ensures Docker Compose stack is clean before and after all tests.
func TestMain(m *testing.M) {
	// Cleanup any existing stack
	_ = runDockerCompose("down", "-v")

	// Run tests
	exitCode := m.Run()

	// Cleanup after tests
	_ = runDockerCompose("down", "-v")

	os.Exit(exitCode)
}

// TestDockerComposeFullStack validates full stack deployment lifecycle.
func TestDockerComposeFullStack(t *testing.T) {
	// Step 1: Start full stack
	t.Log("Starting Docker Compose stack...")
	err := runDockerCompose("up", "-d")
	require.NoError(t, err, "docker compose up should succeed")

	// Step 2: Wait for health checks
	t.Log("Waiting for health checks...")
	time.Sleep(90 * time.Second) // Increased wait time for TLS endpoints to be fully ready

	// Step 3: Verify all services running
	t.Log("Verifying services status...")
	err = runDockerCompose("ps")
	require.NoError(t, err, "docker compose ps should succeed")

	// Step 4: Validate health endpoints
	tests := []struct {
		name string
		url  string
	}{
		{name: "cipher-im-sqlite livez", url: "https://127.0.0.1:9090/admin/v1/livez"},
		{name: "cipher-im-pg-1 livez", url: "https://127.0.0.1:9091/admin/v1/livez"},
		{name: "cipher-im-pg-2 livez", url: "https://127.0.0.1:9092/admin/v1/livez"},
	}

	client := createHTTPSClient()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Get(tt.url)
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

	// Step 5: Cleanup
	t.Log("Cleaning up Docker Compose stack...")
	err = runDockerCompose("down", "-v")
	require.NoError(t, err, "docker compose down should succeed")

	// Step 6: Verify cleanup
	err = runDockerCompose("ps")
	require.NoError(t, err, "docker compose ps should succeed after cleanup")
}

// TestAllInstancesHealthy validates all 3 cipher-im instances are healthy.
func TestAllInstancesHealthy(t *testing.T) {
	// Start stack
	err := runDockerCompose("up", "-d")
	require.NoError(t, err, "docker compose up should succeed")
	defer func() { _ = runDockerCompose("down", "-v") }()

	// Wait for health checks
	time.Sleep(90 * time.Second) // Increased wait time for TLS endpoints

	tests := []struct {
		name          string
		adminPort     int
		containerName string
	}{
		{name: "cipher-im-sqlite", adminPort: 9090, containerName: "cipher-im-sqlite"},
		{name: "cipher-im-pg-1", adminPort: 9091, containerName: "cipher-im-pg-1"},
		{name: "cipher-im-pg-2", adminPort: 9092, containerName: "cipher-im-pg-2"},
	}

	client := createHTTPSClient()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Validate livez endpoint
			livezURL := fmt.Sprintf("https://%s:%d/admin/v1/livez", cryptoutilMagic.IPv4Loopback, tt.adminPort)
			resp, err := client.Get(livezURL)
			require.NoError(t, err, "GET livez should succeed")
			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, http.StatusOK, resp.StatusCode, "livez should return 200 OK")

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err, "reading livez body should succeed")

			var status map[string]string
			err = json.Unmarshal(body, &status)
			require.NoError(t, err, "unmarshaling livez JSON should succeed")

			require.Equal(t, "alive", status["status"], "livez status should be 'alive'")

			// Validate readyz endpoint
			readyzURL := fmt.Sprintf("https://%s:%d/admin/v1/readyz", cryptoutilMagic.IPv4Loopback, tt.adminPort)
			resp, err = client.Get(readyzURL)
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

// createHTTPSClient creates HTTP client with TLS verification disabled for testing.
func createHTTPSClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // Test environment only
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   httpClientTimeout,
	}
}

// runDockerCompose executes docker compose command from cmd/cipher-im directory.
func runDockerCompose(args ...string) error {
	composeDir := filepath.Join("..", "..", "..", "..", "..", "..", "cmd", "cipher-im")

	cmdArgs := append([]string{"compose"}, args...)
	cmd := exec.Command("docker", cmdArgs...)
	cmd.Dir = composeDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
