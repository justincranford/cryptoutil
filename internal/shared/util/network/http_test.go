// Copyright (c) 2025 Justin Cranford

package network

import (
	"context"
	http "net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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
		case "/livez":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		case "/readyz":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		case "/shutdown":
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
		time.Sleep(500 * time.Millisecond)
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

func TestHTTPResponse_GET(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	statusCode, headers, body, err := HTTPResponse(ctx, http.MethodGet, testHTTPServer.URL, time.Second, true, nil, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, headers)
	require.Equal(t, []byte("OK"), body)
}

func TestHTTPResponse_POST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	statusCode, headers, body, err := HTTPResponse(ctx, http.MethodPost, testHTTPServer.URL, time.Second, true, nil, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, statusCode)
	require.NotNil(t, headers)
	require.Equal(t, []byte("Created"), body)
}

func TestHTTPResponse_NoFollowRedirects(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Without following redirects, should get 302.
	statusCode, _, _, err := HTTPResponse(ctx, http.MethodGet, testRedirectServer.URL, time.Second, false, nil, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusFound, statusCode)
}

func TestHTTPResponse_FollowRedirects(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// With following redirects, should get 200.
	statusCode, _, body, err := HTTPResponse(ctx, http.MethodGet, testRedirectServer.URL, time.Second, true, nil, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, []byte("Final"), body)
}

func TestHTTPResponse_Timeout(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use very short timeout (server sleeps 500ms, timeout is 50ms).
	_, _, _, err := HTTPResponse(ctx, http.MethodGet, testSlowServer.URL, 50*time.Millisecond, true, nil, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context deadline exceeded")
}

func TestHTTPResponse_InvalidURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	_, _, _, err := HTTPResponse(ctx, http.MethodGet, "not-a-valid-url", time.Second, true, nil, false)
	require.Error(t, err)
}

func TestHTTPResponse_InvalidMethod(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use an invalid HTTP method with null byte to trigger NewRequestWithContext error.
	// Per RFC 7230, method names are tokens and cannot contain control characters.
	invalidMethod := "GET\x00INVALID"

	_, _, _, err := HTTPResponse(ctx, invalidMethod, "http://example.com", time.Second, true, nil, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create")
}

func TestHTTPResponse_ConnectionRefused(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Try to connect to a port that is not listening.
	_, _, _, err := HTTPResponse(ctx, http.MethodGet, "http://127.0.0.1:1", time.Second, true, nil, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to make GET request")
}

func TestHTTPGetLivez(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	statusCode, headers, body, err := HTTPGetLivez(ctx, testHTTPServer.URL, "", time.Second, nil, true)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, headers)
	require.Equal(t, []byte("OK"), body)
}

func TestHTTPGetLivez_Error(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Invalid URL should trigger HTTPResponse error, which HTTPGetLivez wraps.
	_, _, _, err := HTTPGetLivez(ctx, "http://127.0.0.1:1", "", 100*time.Millisecond, nil, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get")
}

func TestHTTPGetReadyz(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	statusCode, headers, body, err := HTTPGetReadyz(ctx, testHTTPServer.URL, "", time.Second, nil, true)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, headers)
	require.Equal(t, []byte("OK"), body)
}

func TestHTTPGetReadyz_Error(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Invalid URL should trigger HTTPResponse error, which HTTPGetReadyz wraps.
	_, _, _, err := HTTPGetReadyz(ctx, "http://127.0.0.1:1", "", 100*time.Millisecond, nil, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get")
}

func TestHTTPPostShutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	statusCode, headers, body, err := HTTPPostShutdown(ctx, testHTTPServer.URL, "", time.Second, nil, true)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, headers)
	require.Equal(t, []byte("Shutting down"), body)
}

func TestHTTPPostShutdown_Error(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Invalid URL should trigger HTTPResponse error, which HTTPPostShutdown wraps.
	_, _, _, err := HTTPPostShutdown(ctx, "http://127.0.0.1:1", "", 100*time.Millisecond, nil, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to post")
}

func TestHTTPResponse_HTTPS_InsecureSkipVerify(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use insecureSkipVerify to bypass TLS certificate verification.
	statusCode, _, body, err := HTTPResponse(ctx, http.MethodGet, testHTTPSServer.URL, time.Second, true, nil, true)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, []byte("HTTPS OK"), body)
}

func TestHTTPResponse_HTTPS_WithRootCA(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a cert pool with the server's certificate.
	transport, ok := testHTTPSServer.Client().Transport.(*http.Transport)
	require.True(t, ok, "expected *http.Transport")

	rootCAs := transport.TLSClientConfig.RootCAs

	statusCode, _, body, err := HTTPResponse(ctx, http.MethodGet, testHTTPSServer.URL, time.Second, true, rootCAs, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, []byte("HTTPS OK"), body)
}

func TestHTTPResponse_HTTPS_SystemDefaults(t *testing.T) {
	t.Parallel()

	// Test the system defaults path (rootCAsPool == nil && !insecureSkipVerify).
	// Use a real HTTPS endpoint (Google DNS) to test system CA verification.
	ctx := context.Background()

	// This tests the "rootCAsPool == nil && !insecureSkipVerify" path.
	statusCode, _, _, err := HTTPResponse(ctx, http.MethodGet, "https://dns.google", 5*time.Second, true, nil, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
}

func TestHTTPResponse_NoTimeout(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Test with 0 timeout (no timeout).
	statusCode, _, body, err := HTTPResponse(ctx, http.MethodGet, testHTTPServer.URL, 0, true, nil, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, []byte("OK"), body)
}

func TestHTTPResponse_BodyCloseError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// This will exercise the defer function's Close() call.
	// The close error path (fmt.Printf) is hard to test directly since
	// httptest doesn't fail on Close(), but this ensures the defer executes.
	statusCode, _, body, err := HTTPResponse(ctx, http.MethodGet, testHTTPServer.URL, time.Second, true, nil, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, []byte("OK"), body)
}

func TestHTTPResponse_ReadBodyError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use a very short timeout to trigger context cancellation during body read.
	// Note: testSlowServer sleeps 500ms, so 1ns timeout will cancel during read.
	_, _, _, err := HTTPResponse(ctx, http.MethodGet, testSlowServer.URL, 1*time.Nanosecond, true, nil, false)
	require.Error(t, err)
}
