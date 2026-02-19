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
	"strconv"
	"strings"
	"testing"
	"time"

	cryptoutilKmsClient "cryptoutil/internal/apps/sm/kms/client"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilNetwork "cryptoutil/internal/shared/util/network"

	"github.com/stretchr/testify/require"
)

var (
	testSettings                   = cryptoutilAppsTemplateServiceConfig.RequireNewForTest("application_test")
	startServerListenerApplication *ServerApplicationListener
	testServerPublicURL            string
	testServerPrivateURL           string
)

func TestMain(m *testing.M) {
	var err error

	startServerListenerApplication, err = StartServerListenerApplication(testSettings)
	if err != nil {
		log.Fatalf("failed to start server application: %v", err)
	}

	go startServerListenerApplication.StartFunction()

	defer startServerListenerApplication.ShutdownFunction()

	// Build URLs using actual assigned ports
	testServerPublicURL = testSettings.BindPublicProtocol + "://" + testSettings.BindPublicAddress + ":" + strconv.Itoa(int(startServerListenerApplication.ActualPublicPort))
	testServerPrivateURL = testSettings.BindPrivateProtocol + "://" + testSettings.BindPrivateAddress + ":" + strconv.Itoa(int(startServerListenerApplication.ActualPrivatePort))

	cryptoutilKmsClient.WaitUntilReady(&testServerPrivateURL, cryptoutilSharedMagic.TimeoutTestServerReady, cryptoutilSharedMagic.TimeoutTestServerReadyRetryDelay, startServerListenerApplication.PrivateTLSServer.RootCAsPool)

	exitCode := m.Run()
	if exitCode != 0 {
		fmt.Printf("Tests failed with exit code %d\n", exitCode)
	}
}

func TestHttpGetTraceHead(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		method         string
		url            string
		tlsRootCAs     *x509.CertPool
		expectedStatus int
		expectError    bool
	}{
		{name: "Swagger UI root", method: "GET", url: testServerPublicURL + "/ui/swagger", tlsRootCAs: startServerListenerApplication.PublicTLSServer.RootCAsPool, expectedStatus: http.StatusMovedPermanently, expectError: false},
		{name: "Swagger UI index.html", method: "GET", url: testServerPublicURL + "/ui/swagger/index.html", tlsRootCAs: startServerListenerApplication.PublicTLSServer.RootCAsPool, expectedStatus: http.StatusOK, expectError: false},
		{name: "OpenAPI Spec", method: "GET", url: testServerPublicURL + "/ui/swagger/doc.json", tlsRootCAs: startServerListenerApplication.PublicTLSServer.RootCAsPool, expectedStatus: http.StatusOK, expectError: false},
		{name: "GET Elastic Keys", method: "GET", url: testServerPublicURL + testSettings.PublicServiceAPIContextPath + "/elastickeys", tlsRootCAs: startServerListenerApplication.PublicTLSServer.RootCAsPool, expectedStatus: http.StatusOK, expectError: false},

		{name: "HEAD Elastic Keys", method: "HEAD", url: testServerPublicURL + testSettings.PublicServiceAPIContextPath + "/elastickeys", tlsRootCAs: startServerListenerApplication.PublicTLSServer.RootCAsPool, expectedStatus: http.StatusMethodNotAllowed, expectError: false},
		{name: "TRACE Elastic Keys", method: "TRACE", url: testServerPublicURL + testSettings.PublicServiceAPIContextPath + "/elastickeys", tlsRootCAs: startServerListenerApplication.PublicTLSServer.RootCAsPool, expectedStatus: http.StatusMethodNotAllowed, expectError: false},

		{name: "GET Non-existent endpoint", method: "GET", url: testServerPublicURL + "/nonexistent", tlsRootCAs: startServerListenerApplication.PublicTLSServer.RootCAsPool, expectedStatus: http.StatusBadRequest, expectError: false},
		{name: "GET Service API without TLS", method: "GET", url: "http://" + testSettings.BindPublicAddress + ":" + strconv.Itoa(int(startServerListenerApplication.ActualPublicPort)) + testSettings.PublicServiceAPIContextPath + "/elastickeys", tlsRootCAs: nil, expectedStatus: 0, expectError: true},
		{name: "GET Service API with TLS", method: "GET", url: testSettings.BindPublicProtocol + "://" + testSettings.BindPublicAddress + ":" + strconv.Itoa(int(startServerListenerApplication.ActualPublicPort)) + testSettings.PublicServiceAPIContextPath + "/elastickeys", tlsRootCAs: startServerListenerApplication.PublicTLSServer.RootCAsPool, expectedStatus: http.StatusOK, expectError: false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			statusCode, headers, body, err := cryptoutilSharedUtilNetwork.HTTPResponse(context.Background(), tc.method, tc.url, 10*time.Second, false, tc.tlsRootCAs, false)
			if tc.expectError {
				require.Error(t, err, "expected request to fail")
				return //nolint:nlreturn // gofumpt removes blank line required by nlreturn linter
			}

			require.NoError(t, err, "failed to get response")
			require.NotNil(t, body, "response body should not be nil")
			require.NotNil(t, headers, "response headers should not be nil")

			// Check status code
			require.Equal(t, tc.expectedStatus, statusCode)

			var contentString string
			if body != nil {
				contentString = strings.ReplaceAll(string(body), "\n", " ")
			}

			if err == nil {
				t.Logf("PASS: %s, Contents: %s", tc.url, contentString)
			} else {
				require.Fail(t, fmt.Sprintf("FAILED: %s, Contents: %s, Error: %v", tc.url, contentString, err))
			}
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		url               string
		method            string
		isHTTPS           bool
		isBrowserPath     bool
		isLogoutEndpoint  bool
		tlsRootCAs        *x509.CertPool
		expectedHeaders   map[string]string
		unexpectedHeaders []string
	}{
		{
			name:          "Browser API HTTPS - Standard endpoint",
			url:           testServerPublicURL + testSettings.PublicBrowserAPIContextPath + "/elastickeys",
			method:        "GET",
			isHTTPS:       strings.HasPrefix(testServerPublicURL, "https://"),
			isBrowserPath: true,
			tlsRootCAs:    startServerListenerApplication.PublicTLSServer.RootCAsPool,
			expectedHeaders: map[string]string{
				"X-Content-Type-Options":            "nosniff",
				"Referrer-Policy":                   "strict-origin-when-cross-origin",
				"Strict-Transport-Security":         "max-age=86400; includeSubDomains",
				"Permissions-Policy":                "camera=(), microphone=(), geolocation=(), payment=(), usb=(), accelerometer=(), gyroscope=(), magnetometer=()",
				"Cross-Origin-Opener-Policy":        "same-origin",
				"Cross-Origin-Embedder-Policy":      "require-corp",
				"Cross-Origin-Resource-Policy":      "same-origin",
				"X-Permitted-Cross-Domain-Policies": "none",
			},
			unexpectedHeaders: []string{"Clear-Site-Data"},
		},
		{
			name:          "Service API HTTPS - Standard endpoint",
			url:           testServerPublicURL + testSettings.PublicServiceAPIContextPath + "/elastickeys",
			method:        "GET",
			isHTTPS:       strings.HasPrefix(testServerPublicURL, "https://"),
			isBrowserPath: false,
			tlsRootCAs:    startServerListenerApplication.PublicTLSServer.RootCAsPool,
			expectedHeaders: map[string]string{
				// Service API has minimal headers since Helmet and our security middleware are skipped
				"X-Content-Type-Options":    "nosniff",
				"Referrer-Policy":           "strict-origin-when-cross-origin",
				"Strict-Transport-Security": "max-age=86400; includeSubDomains",
			},
			unexpectedHeaders: []string{
				"Permissions-Policy",
				"Cross-Origin-Opener-Policy",
				"Cross-Origin-Embedder-Policy",
				"Cross-Origin-Resource-Policy",
				"Clear-Site-Data",
			},
		},
		// Note: We cannot easily test POST /logout without authentication setup in this test
		// The logout functionality would require CSRF tokens and authentication
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			statusCode, headers, body, err := cryptoutilSharedUtilNetwork.HTTPResponse(context.Background(), "GET", tc.url, 10*time.Second, false, tc.tlsRootCAs, false)
			require.NotNil(t, body, "response body should not be nil")
			require.NotNil(t, headers, "response headers should not be nil")
			require.NoError(t, err, "failed to get response headers")
			require.Equal(t, http.StatusOK, statusCode, "should return 200 OK")

			// Check expected headers are present and have correct values
			for expectedHeader, expectedValue := range tc.expectedHeaders {
				actualValue := headers.Get(expectedHeader)
				if expectedValue != "" {
					require.Equal(t, expectedValue, actualValue,
						"Header %s should have value %s but got %s", expectedHeader, expectedValue, actualValue)
				} else {
					require.NotEmpty(t, actualValue, "Header %s should be present", expectedHeader)
				}
			}

			// Check unexpected headers are not present
			for _, unexpectedHeader := range tc.unexpectedHeaders {
				actualValue := headers.Get(unexpectedHeader)
				require.Empty(t, actualValue, "Header %s should not be present but got %s", unexpectedHeader, actualValue)
			}

			// HTTPS-specific checks
			if tc.isHTTPS {
				hstsValue := headers.Get("Strict-Transport-Security")
				if tc.isBrowserPath {
					require.NotEmpty(t, hstsValue, "HSTS header should be present on HTTPS browser requests")

					if testSettings.DevMode {
						require.Contains(t, hstsValue, "max-age=86400", "HSTS should use shorter duration in dev mode")
					} else {
						require.Contains(t, hstsValue, "max-age=31536000", "HSTS should use 1 year duration in production")
						require.Contains(t, hstsValue, "preload", "HSTS should include preload in production")
					}

					require.Contains(t, hstsValue, "includeSubDomains", "HSTS should include subdomains")
				}
			}

			t.Logf("âœ“ Security headers validated for %s", tc.name)
		})
	}
}

func TestHealthChecks(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		endpoint       string
		getResponse    func(*string, *x509.CertPool) (int, http.Header, []byte, error)
		expectedStatus int
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name:     "Liveness Check (" + cryptoutilSharedMagic.PrivateAdminLivezRequestPath + ")",
			endpoint: cryptoutilSharedMagic.PrivateAdminLivezRequestPath,
			getResponse: func(baseURL *string, rootCAsPool *x509.CertPool) (int, http.Header, []byte, error) {
				return cryptoutilSharedUtilNetwork.HTTPGetLivez(context.Background(), *baseURL, cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath, 2*time.Second, rootCAsPool, false)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				t.Helper()

				var response map[string]any

				err := json.Unmarshal(body, &response)
				require.NoError(t, err, "should return valid JSON")

				// Liveness should always be "ok"
				require.Equal(t, "ok", response["status"], "liveness status should be 'ok'")
				require.Equal(t, "liveness", response["probe"], "probe should be 'liveness'")
				require.NotEmpty(t, response["timestamp"], "should include timestamp")
				require.Equal(t, "cryptoutil", response["service"], "service name should be 'cryptoutil'")

				// Liveness should not include detailed checks
				require.NotContains(t, response, "database", "liveness should not include database checks")
				require.NotContains(t, response, "memory", "liveness should not include memory checks")
				require.NotContains(t, response, "dependencies", "liveness should not include dependency checks")
			},
		},
		{
			name:     "Readiness Check (" + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath + ")",
			endpoint: cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			getResponse: func(baseURL *string, rootCAsPool *x509.CertPool) (int, http.Header, []byte, error) {
				return cryptoutilSharedUtilNetwork.HTTPGetReadyz(context.Background(), *baseURL, cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath, 2*time.Second, rootCAsPool, false)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				t.Helper()

				var response map[string]any

				err := json.Unmarshal(body, &response)
				require.NoError(t, err, "should return valid JSON")

				// Readiness should be "ok" in healthy state
				require.Equal(t, "ok", response["status"], "readiness status should be 'ok'")
				require.Equal(t, "readiness", response["probe"], "probe should be 'readiness'")
				require.NotEmpty(t, response["timestamp"], "should include timestamp")
				require.Equal(t, "cryptoutil", response["service"], "service name should be 'cryptoutil'")

				// Readiness should include detailed checks
				require.Contains(t, response, "database", "readiness should include database checks")
				require.Contains(t, response, "memory", "readiness should include memory checks")
				require.Contains(t, response, "dependencies", "readiness should include dependency checks")

				// Validate database structure
				dbStatus, ok := response["database"].(map[string]any)
				require.True(t, ok, "database should be an object")
				require.Contains(t, dbStatus, "status", "database should have status")
				require.Contains(t, dbStatus, "db_type", "database should have db_type")

				// Validate memory structure
				memStatus, ok := response["memory"].(map[string]any)
				require.True(t, ok, "memory should be an object")
				require.Equal(t, "ok", memStatus["status"], "memory status should be 'ok'")
				require.Contains(t, memStatus, "heap_alloc", "memory should include heap_alloc")
				require.Contains(t, memStatus, "num_goroutines", "memory should include num_goroutines")

				// Validate dependencies structure
				depsStatus, ok := response["dependencies"].(map[string]any)
				require.True(t, ok, "dependencies should be an object")
				require.Contains(t, depsStatus, "status", "dependencies should have status")
				require.Contains(t, depsStatus, "services", "dependencies should have services")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Increase timeout for health check requests to prevent flakiness
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Create HTTP client with timeout context
			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						RootCAs: startServerListenerApplication.PrivateTLSServer.RootCAsPool,
					},
				},
				Timeout: 5 * time.Second,
			}

			req, err := http.NewRequestWithContext(ctx, "GET", testServerPrivateURL+tc.endpoint, nil)
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err, "should successfully get response from %s", tc.endpoint)

			defer func() {
				require.NoError(t, resp.Body.Close())
			}()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var response map[string]any

			err = json.Unmarshal(body, &response)
			require.NoError(t, err)

			require.Equal(t, tc.expectedStatus, resp.StatusCode, "should return expected status code")
			require.NotNil(t, response, "response body should not be nil")

			tc.validateBody(t, body)

			t.Logf("âœ“ Health check %s validation passed", tc.endpoint)
		})
	}
}

func TestSendServerListenerLivenessCheck(t *testing.T) {
	t.Parallel()
	// Update test settings to use the actual assigned port for the liveness check
	testSettingsForLiveness := *testSettings
	testSettingsForLiveness.BindPrivatePort = startServerListenerApplication.ActualPrivatePort

	body, err := SendServerListenerLivenessCheck(&testSettingsForLiveness)
	require.NoError(t, err, "SendServerListenerLivenessCheck should not return an error")
	require.NotNil(t, body, "response body should not be nil")
	require.NotEmpty(t, body, "response body should not be empty")
	t.Logf("Liveness check response: %s", body)

	// Parse the JSON response
	var response map[string]any

	err = json.Unmarshal(body, &response)
	require.NoError(t, err, "should return valid JSON")

	// Validate liveness response structure
	require.Equal(t, "ok", response["status"], "liveness status should be 'ok'")
	require.Equal(t, "liveness", response["probe"], "probe should be 'liveness'")
	require.NotEmpty(t, response["timestamp"], "should include timestamp")
	require.Equal(t, "cryptoutil", response["service"], "service name should be 'cryptoutil'")

	// Liveness should not include detailed checks
	require.NotContains(t, response, "database", "liveness should not include database checks")
	require.NotContains(t, response, "memory", "liveness should not include memory checks")
	require.NotContains(t, response, "dependencies", "liveness should not include dependency checks")

	t.Logf("âœ“ SendServerListenerLivenessCheck validation passed")
}
