// Copyright (c) 2025 Justin Cranford

package network

import (
	"context"
	"crypto/x509"
	http "net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const testRootPath = "/"

var (
	// Shared test servers for the entire package.
	testHTTPServer     *httptest.Server
	testHTTPSServer    *httptest.Server
	testSlowServer     *httptest.Server
	testRedirectServer *httptest.Server
)

func TestMain(m *testing.M) {
	// Create shared HTTP server.
	testHTTPServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case testRootPath:
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte("Created"))
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			}
		case cryptoutilSharedMagic.PrivateAdminLivezRequestPath:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		case cryptoutilSharedMagic.PrivateAdminReadyzRequestPath:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		case cryptoutilSharedMagic.PrivateAdminShutdownRequestPath:
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("Shutting down"))
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer testHTTPServer.Close()

	// Create shared HTTPS server.
	testHTTPSServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("HTTPS OK"))
	}))
	defer testHTTPSServer.Close()

	// Create shared slow server for timeout tests.
	testSlowServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(cryptoutilSharedMagic.TestDefaultRateLimitServiceIP * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer testSlowServer.Close()

	// Create shared redirect server.
	testRedirectServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == testRootPath {
			http.Redirect(w, r, "/final", http.StatusFound)

			return
		}

		switch r.URL.Path {
		case "/final":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Final"))
		case "/redirected":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Redirected"))
		}
	}))
	defer testRedirectServer.Close()

	// Run all tests.
	exitCode := m.Run()

	os.Exit(exitCode)
}

func TestHTTPResponse_HappyPaths(t *testing.T) {
	t.Parallel()

	// Pre-extract rootCAs from testHTTPSServer for the WithRootCA case.
	transport, ok := testHTTPSServer.Client().Transport.(*http.Transport)
	require.True(t, ok, "expected *http.Transport")

	httpsRootCAs := transport.TLSClientConfig.RootCAs

	tests := []struct {
		name               string
		method             string
		url                string
		timeout            time.Duration
		followRedirects    bool
		rootCAs            *x509.CertPool
		insecureSkipVerify bool
		wantStatus         int
		wantBody           []byte
	}{
		{name: "GET", method: http.MethodGet, url: testHTTPServer.URL, timeout: time.Second, followRedirects: true, wantStatus: http.StatusOK, wantBody: []byte("OK")},
		{name: "POST", method: http.MethodPost, url: testHTTPServer.URL, timeout: time.Second, followRedirects: true, wantStatus: http.StatusCreated, wantBody: []byte("Created")},
		{name: "no follow redirects", method: http.MethodGet, url: testRedirectServer.URL, timeout: time.Second, wantStatus: http.StatusFound},
		{name: "follow redirects", method: http.MethodGet, url: testRedirectServer.URL, timeout: time.Second, followRedirects: true, wantStatus: http.StatusOK, wantBody: []byte("Final")},
		{name: "HTTPS insecure skip verify", method: http.MethodGet, url: testHTTPSServer.URL, timeout: time.Second, followRedirects: true, insecureSkipVerify: true, wantStatus: http.StatusOK, wantBody: []byte("HTTPS OK")},
		{name: "HTTPS with root CA", method: http.MethodGet, url: testHTTPSServer.URL, timeout: time.Second, followRedirects: true, rootCAs: httpsRootCAs, wantStatus: http.StatusOK, wantBody: []byte("HTTPS OK")},
		{name: "HTTPS system defaults", method: http.MethodGet, url: "https://dns.google", timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second, followRedirects: true, wantStatus: http.StatusOK},
		{name: "no timeout", method: http.MethodGet, url: testHTTPServer.URL, timeout: 0, followRedirects: true, wantStatus: http.StatusOK, wantBody: []byte("OK")},
		{name: "body close exercises defer", method: http.MethodGet, url: testHTTPServer.URL, timeout: time.Second, followRedirects: true, wantStatus: http.StatusOK, wantBody: []byte("OK")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			statusCode, headers, body, err := HTTPResponse(ctx, tc.method, tc.url, tc.timeout, tc.followRedirects, tc.rootCAs, tc.insecureSkipVerify)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatus, statusCode)
			require.NotNil(t, headers)

			if tc.wantBody != nil {
				require.Equal(t, tc.wantBody, body)
			}
		})
	}
}

func TestHTTPResponse_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		method       string
		url          string
		timeout      time.Duration
		wantContains string
	}{
		{name: "timeout", method: http.MethodGet, url: testSlowServer.URL, timeout: cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond, wantContains: "context deadline exceeded"},
		{name: "invalid URL", method: http.MethodGet, url: "not-a-valid-url", timeout: time.Second},
		{name: "invalid method", method: "GET\x00INVALID", url: "http://example.com", timeout: time.Second, wantContains: "failed to create"},
		{name: "connection refused", method: http.MethodGet, url: "http://127.0.0.1:1", timeout: time.Second, wantContains: "failed to make GET request"},
		{name: "read body error", method: http.MethodGet, url: testSlowServer.URL, timeout: time.Nanosecond},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			_, _, _, err := HTTPResponse(ctx, tc.method, tc.url, tc.timeout, true, nil, false)
			require.Error(t, err)

			if tc.wantContains != "" {
				require.Contains(t, err.Error(), tc.wantContains)
			}
		})
	}
}

type healthEndpointFn func(ctx context.Context, baseURL, adminContextPath string, timeout time.Duration, rootCAsPool *x509.CertPool, insecureSkipVerify bool) (int, http.Header, []byte, error)

func TestHealthEndpoints_HappyPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		fn         healthEndpointFn
		wantStatus int
		wantBody   []byte
	}{
		{name: "livez", fn: HTTPGetLivez, wantStatus: http.StatusOK, wantBody: []byte("OK")},
		{name: "readyz", fn: HTTPGetReadyz, wantStatus: http.StatusOK, wantBody: []byte("OK")},
		{name: "shutdown", fn: HTTPPostShutdown, wantStatus: http.StatusOK, wantBody: []byte("Shutting down")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			statusCode, headers, body, err := tc.fn(ctx, testHTTPServer.URL, "", time.Second, nil, true)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatus, statusCode)
			require.NotNil(t, headers)
			require.Equal(t, tc.wantBody, body)
		})
	}
}

func TestHealthEndpoints_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		fn           healthEndpointFn
		wantContains string
	}{
		{name: "livez", fn: HTTPGetLivez, wantContains: "failed to get"},
		{name: "readyz", fn: HTTPGetReadyz, wantContains: "failed to get"},
		{name: "shutdown", fn: HTTPPostShutdown, wantContains: "failed to post"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			_, _, _, err := tc.fn(ctx, "http://127.0.0.1:1", "", cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond, nil, false)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantContains)
		})
	}
}
