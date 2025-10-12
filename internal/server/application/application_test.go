package application

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	cryptoutilClient "cryptoutil/internal/client"
	cryptoutilConfig "cryptoutil/internal/common/config"

	"github.com/stretchr/testify/require"
)

var (
	testSettings                   = cryptoutilConfig.RequireNewForTest("application_test")
	startServerListenerApplication *ServerApplicationListener
	testServerPublicURL            string
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
	testServerPrivateURL := testSettings.BindPrivateProtocol + "://" + testSettings.BindPrivateAddress + ":" + strconv.Itoa(int(startServerListenerApplication.ActualPrivatePort))

	cryptoutilClient.WaitUntilReady(&testServerPrivateURL, 3*time.Second, 100*time.Millisecond, startServerListenerApplication.PrivateTLSServer.RootCAsPool)

	exitCode := m.Run()
	if exitCode != 0 {
		fmt.Printf("Tests failed with exit code %d\n", exitCode)
	}
}

func TestHttpGetTraceHead(t *testing.T) {
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
			body, headers, err := httpResponse(t, tc.method, tc.expectedStatus, tc.url, tc.tlsRootCAs)
			if tc.expectError {
				require.Error(t, err, "expected request to fail")
				return
			}
			require.NotNil(t, body, "response body should not be nil")
			require.NotNil(t, headers, "response headers should not be nil")
			require.NoError(t, err, "failed to get response headers")
			var contentString string
			if body != nil {
				contentString = strings.Replace(string(body), "\n", " ", -1)
			}
			if err == nil {
				t.Logf("PASS: %s, Contents: %s", tc.url, contentString)
			} else {
				t.Errorf("FAILED: %s, Contents: %s, Error: %v", tc.url, contentString, err)
			}
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
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
			body, headers, err := httpResponse(t, "GET", http.StatusOK, tc.url, tc.tlsRootCAs)
			require.NotNil(t, body, "response body should not be nil")
			require.NotNil(t, headers, "response headers should not be nil")
			require.NoError(t, err, "failed to get response headers")

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

			t.Logf("✓ Security headers validated for %s", tc.name)
		})
	}
}

func httpResponse(t *testing.T, httpMethod string, expectedStatusCode int, url string, rootCAsPool *x509.CertPool) ([]byte, http.Header, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), httpMethod, url, nil)
	require.NoError(t, err, "failed to create %s request", httpMethod)
	req.Header.Set("Accept", "*/*")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}
	if strings.HasPrefix(url, "https://") {
		transport := &http.Transport{}
		if rootCAsPool != nil {
			transport.TLSClientConfig = &tls.Config{
				RootCAs:    rootCAsPool,
				MinVersion: tls.VersionTLS12,
			}
		} else {
			transport.TLSClientConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		}
		client.Transport = transport
	}

	resp, err := client.Do(req)
	if err != nil {
		if expectedStatusCode == 0 {
			return nil, nil, fmt.Errorf("expected error occurred: %w", err)
		}
		require.NoError(t, err, "failed to make %s request", httpMethod)
	}
	defer func() {
		if resp != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				t.Errorf("Warning: failed to close response body: %v", closeErr)
			}
		}
	}()

	if resp == nil {
		return nil, nil, fmt.Errorf("no response received")
	}

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "HTTP Status code: "+strconv.Itoa(resp.StatusCode)+", failed to read error response body")
	if expectedStatusCode != 0 && resp.StatusCode != expectedStatusCode {
		return nil, nil, fmt.Errorf("HTTP Status code: %d, error response body: %v", resp.StatusCode, string(body))
	}
	t.Logf("HTTP Status code: %d, response headers count: %d, response body: %d bytes", resp.StatusCode, len(resp.Header), len(body))
	return body, resp.Header, nil
}
