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
		if r.URL.Path == "/" {
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
		if r.URL.Path == "/" {
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
	rootCAs := server.Client().Transport.(*http.Transport).TLSClientConfig.RootCAs

	statusCode, _, body, err := HTTPResponse(ctx, http.MethodGet, server.URL, time.Second, true, rootCAs, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, []byte("HTTPS OK"), body)
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
