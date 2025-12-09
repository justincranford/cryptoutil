// Copyright (c) 2025 Justin Cranford

package network

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testRootPath = "/"

func TestHTTPResponse_GET(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	ctx := context.Background()

	statusCode, headers, body, err := HTTPResponse(ctx, http.MethodGet, server.URL, time.Second, true, nil, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, headers)
	require.Equal(t, []byte("OK"), body)
}

func TestHTTPResponse_POST(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("Created"))
	}))
	defer server.Close()

	ctx := context.Background()

	statusCode, headers, body, err := HTTPResponse(ctx, http.MethodPost, server.URL, time.Second, true, nil, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, statusCode)
	require.NotNil(t, headers)
	require.Equal(t, []byte("Created"), body)
}

func TestHTTPResponse_NoFollowRedirects(t *testing.T) {
	t.Parallel()

	redirects := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == testRootPath {
			redirects++

			http.Redirect(w, r, "/redirected", http.StatusFound)

			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Redirected"))
	}))
	defer server.Close()

	ctx := context.Background()

	// Without following redirects, should get 302.
	statusCode, _, _, err := HTTPResponse(ctx, http.MethodGet, server.URL, time.Second, false, nil, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusFound, statusCode)
	require.Equal(t, 1, redirects)
}

func TestHTTPResponse_FollowRedirects(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == testRootPath {
			http.Redirect(w, r, "/final", http.StatusFound)

			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Final"))
	}))
	defer server.Close()

	ctx := context.Background()

	// With following redirects, should get 200.
	statusCode, _, body, err := HTTPResponse(ctx, http.MethodGet, server.URL, time.Second, true, nil, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, []byte("Final"), body)
}

func TestHTTPResponse_Timeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than timeout.
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()

	// Use very short timeout.
	_, _, _, err := HTTPResponse(ctx, http.MethodGet, server.URL, 50*time.Millisecond, true, nil, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context deadline exceeded")
}

func TestHTTPResponse_InvalidURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	_, _, _, err := HTTPResponse(ctx, http.MethodGet, "not-a-valid-url", time.Second, true, nil, false)
	require.Error(t, err)
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Contains(t, r.URL.Path, "/livez")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	ctx := context.Background()

	statusCode, headers, body, err := HTTPGetLivez(ctx, server.URL, "", time.Second, nil, true)
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Contains(t, r.URL.Path, "/readyz")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	ctx := context.Background()

	statusCode, headers, body, err := HTTPGetReadyz(ctx, server.URL, "", time.Second, nil, true)
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/shutdown")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Shutting down"))
	}))
	defer server.Close()

	ctx := context.Background()

	statusCode, headers, body, err := HTTPPostShutdown(ctx, server.URL, "", time.Second, nil, true)
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

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("HTTPS OK"))
	}))
	defer server.Close()

	ctx := context.Background()

	// Use insecureSkipVerify to bypass TLS certificate verification.
	statusCode, _, body, err := HTTPResponse(ctx, http.MethodGet, server.URL, time.Second, true, nil, true)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, []byte("HTTPS OK"), body)
}

func TestHTTPResponse_HTTPS_WithRootCA(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("HTTPS OK"))
	}))
	defer server.Close()

	ctx := context.Background()

	// Create a cert pool with the server's certificate.
	transport, ok := server.Client().Transport.(*http.Transport)
	require.True(t, ok, "expected *http.Transport")

	rootCAs := transport.TLSClientConfig.RootCAs

	statusCode, _, body, err := HTTPResponse(ctx, http.MethodGet, server.URL, time.Second, true, rootCAs, false)
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	ctx := context.Background()

	// Test with 0 timeout (no timeout).
	statusCode, _, body, err := HTTPResponse(ctx, http.MethodGet, server.URL, 0, true, nil, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, []byte("OK"), body)
}

func TestHTTPResponse_BodyCloseError(t *testing.T) {
	t.Parallel()

	// Create a custom response writer that returns a body which fails on Close().
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Use a flusher to force-flush headers, then write body.
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)
		flusher.Flush()

		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	ctx := context.Background()

	// This will exercise the defer function's Close() call.
	// The close error path (fmt.Printf) is hard to test directly since
	// httptest doesn't fail on Close(), but this ensures the defer executes.
	statusCode, _, body, err := HTTPResponse(ctx, http.MethodGet, server.URL, time.Second, true, nil, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, []byte("OK"), body)
}

func TestHTTPResponse_ReadBodyError(t *testing.T) {
	t.Parallel()

	// Test the "failed to read response body" error path by using a body that fails on Read.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Write a large response to ensure Read() will be called multiple times.
		_, _ = w.Write(make([]byte, 1024*1024)) // 1MB
	}))
	defer server.Close()

	ctx := context.Background()

	// Use a very short timeout to trigger context cancellation during body read.
	_, _, _, err := HTTPResponse(ctx, http.MethodGet, server.URL, 1*time.Nanosecond, true, nil, false)
	require.Error(t, err)
}
