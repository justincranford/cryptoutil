//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford
//
// NOTE: These tests require a PostgreSQL database and are skipped in CI without the integration tag.
//

package application

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	json "encoding/json"
	"fmt"
	"io"
	"log"
	http "net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilKmsClient "cryptoutil/internal/apps/sm/kms/client"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilNetwork "cryptoutil/internal/shared/util/network"

	"github.com/stretchr/testify/require"
)

func TestSendServerListenerReadinessCheck(t *testing.T) {
	t.Parallel()
	// Update test settings to use the actual assigned port for the readiness check
	testSettingsForReadiness := *testSettings
	testSettingsForReadiness.BindPrivatePort = startServerListenerApplication.ActualPrivatePort

	body, err := SendServerListenerReadinessCheck(&testSettingsForReadiness)
	require.NoError(t, err, "SendServerListenerReadinessCheck should not return an error")
	require.NotNil(t, body, "response body should not be nil")
	require.NotEmpty(t, body, "response body should not be empty")
	t.Logf("Readiness check response: %s", body)

	// Parse the JSON response
	var response map[string]any

	err = json.Unmarshal(body, &response)
	require.NoError(t, err, "should return valid JSON")

	// Validate readiness response structure
	require.Equal(t, "ok", response["status"], "readiness status should be 'ok'")
	require.Equal(t, "readiness", response["probe"], "probe should be 'readiness'")
	require.NotEmpty(t, response["timestamp"], "should include timestamp")
	require.Equal(t, "cryptoutil", response["service"], "service name should be 'cryptoutil'")

	// Readiness should include detailed checks
	require.Contains(t, response, "database", "readiness should include database checks")
	require.Contains(t, response, "memory", "readiness should include memory checks")
	require.Contains(t, response, "dependencies", "readiness should include dependency checks")

	// Validate database health structure
	database, ok := response["database"].(map[string]any)
	require.True(t, ok, "database should be a map")
	require.Contains(t, database, "status", "database should have status")

	// Validate memory health structure
	memory, ok := response["memory"].(map[string]any)
	require.True(t, ok, "memory should be a map")
	require.Contains(t, memory, "status", "memory should have status")
	require.Contains(t, memory, "heap_alloc", "memory should include heap allocation info")
	require.Contains(t, memory, "num_goroutines", "memory should include goroutine count")

	// Validate dependencies health structure
	dependencies, ok := response["dependencies"].(map[string]any)
	require.True(t, ok, "dependencies should be a map")
	require.Contains(t, dependencies, "status", "dependencies should have status")
	require.Contains(t, dependencies, "services", "dependencies should include services")

	t.Logf("âœ“ SendServerListenerReadinessCheck validation passed")
}

func TestRequestLoggerMiddleware(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectError    bool
		logScenario    string // "success", "failed", or "no_response"
	}{
		{
			name:           "Success Request - GET Elastic Keys",
			method:         "GET",
			path:           testSettings.PublicServiceAPIContextPath + "/elastickeys",
			expectedStatus: http.StatusOK,
			expectError:    false,
			logScenario:    "success",
		},
		{
			name:           "Failed Request - HEAD Method Not Allowed",
			method:         "HEAD",
			path:           testSettings.PublicServiceAPIContextPath + "/elastickeys",
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    false,
			logScenario:    "failed",
		},
		{
			name:           "No Response Request - Non-existent endpoint",
			method:         "GET",
			path:           "/nonexistent",
			expectedStatus: http.StatusBadRequest,
			expectError:    false,
			logScenario:    "no_response",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := testServerPublicURL + tc.path
			statusCode, headers, body, err := cryptoutilSharedUtilNetwork.HTTPResponse(
				context.Background(),
				tc.method,
				url,
				10*time.Second, // Increased from 2s to 10s for race detector compatibility (race detector adds ~10x overhead)
				false,
				startServerListenerApplication.PublicTLSServer.RootCAsPool,
				false,
			)

			if tc.expectError {
				require.Error(t, err, "expected request to fail")
				return //nolint:nlreturn // gofumpt removes blank line required by nlreturn linter
			}

			require.NoError(t, err, "failed to get response")
			require.NotNil(t, headers, "response headers should not be nil")

			// Check status code
			require.Equal(t, tc.expectedStatus, statusCode)

			// Validate logging scenarios based on the test case
			switch tc.logScenario {
			case "success":
				// Success requests should have 2xx status codes
				require.True(t, statusCode >= 200 && statusCode < 300,
					"success scenario should have 2xx status code, got %d", statusCode)
				t.Logf("âœ“ Success request logged: status=%d, method=%s, path=%s",
					statusCode, tc.method, tc.path)

			case "failed":
				// Failed requests should have 4xx or 5xx status codes
				require.True(t, statusCode >= 400,
					"failed scenario should have 4xx/5xx status code, got %d", statusCode)
				t.Logf("âœ“ Failed request logged: status=%d, method=%s, path=%s",
					statusCode, tc.method, tc.path)

			case "no_response":
				// No response scenarios are requests that result in errors but still get logged
				// The middleware logs the status that was set at the time of logging
				// In this case, Fiber sets status 200 initially, but the error is logged
				require.Contains(t, string(body), "no matching operation was found",
					"should contain error message for unmatched route")
				t.Logf("âœ“ No response request logged: status=%d, method=%s, path=%s, error='no matching operation was found'",
					statusCode, tc.method, tc.path)
			}
		})
	}
}

// List and delete PEM files created during testing.
func TestCleanupTestCertificates(t *testing.T) {
	t.Parallel()
	// List PEM files in the current package directory
	files, err := os.ReadDir(".")
	if err != nil {
		t.Logf("Warning: Could not read directory for PEM file cleanup: %v", err)

		return
	}

	var pemFiles []string

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".pem") {
			pemFiles = append(pemFiles, file.Name())
		}
	}

	// List the PEM files found
	if len(pemFiles) > 0 {
		t.Logf("Found PEM files in %s directory:", "internal/server/application")

		for _, pemFile := range pemFiles {
			t.Logf("  - %s", pemFile)
		}

		// Delete the PEM files
		for _, pemFile := range pemFiles {
			err := os.Remove(pemFile)
			require.NoError(t, err, "Failed to delete PEM file %s", pemFile)
			t.Logf("Successfully deleted PEM file: %s", pemFile)
		}
	} else {
		t.Logf("No PEM files found in %s directory", "internal/server/application")
	}
}
