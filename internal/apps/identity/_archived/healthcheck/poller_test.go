// Copyright (c) 2025 Justin Cranford

package healthcheck

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPollerPollHealthy(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/health", r.URL.Path)

		resp := Response{
			Status:   cryptoutilSharedMagic.DockerServiceHealthHealthy,
			Database: "ok",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp) //nolint:errcheck // Test HTTP handler - encoding error in test response not critical
	}))
	defer server.Close()

	poller := NewPoller(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second, 3, true) // skipTLSVerify=true for tests
	ctx := context.Background()

	resp, err := poller.Poll(ctx, server.URL+"/health")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, cryptoutilSharedMagic.DockerServiceHealthHealthy, resp.Status)
	require.Equal(t, "ok", resp.Database)
}

func TestPollerPollUnhealthy(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := Response{
			Status: "unhealthy",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp) //nolint:errcheck // Test HTTP handler - encoding error in test response not critical
	}))
	defer server.Close()

	poller := NewPoller(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second, 2, true) // skipTLSVerify=true for tests
	ctx := context.Background()

	resp, err := poller.Poll(ctx, server.URL+"/health")
	require.Error(t, err)
	require.Contains(t, err.Error(), "health check failed")
	require.Nil(t, resp)
}

func TestPollerPollNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	poller := NewPoller(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second, 2, true) // skipTLSVerify=true for tests
	ctx := context.Background()

	resp, err := poller.Poll(ctx, server.URL+"/health")
	require.Error(t, err)
	require.Contains(t, err.Error(), "health check failed")
	require.Nil(t, resp)
}

func TestPollerPollInvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json{{{")) //nolint:errcheck // Test HTTP handler - error not critical for test validation
	}))
	defer server.Close()

	poller := NewPoller(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second, 2, true) // skipTLSVerify=true for tests
	ctx := context.Background()

	resp, err := poller.Poll(ctx, server.URL+"/health")
	require.Error(t, err)
	require.Contains(t, err.Error(), "health check failed")
	require.Nil(t, resp)
}

func TestPollerPollContextCanceled(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)

		resp := Response{Status: cryptoutilSharedMagic.DockerServiceHealthHealthy}
		_ = json.NewEncoder(w).Encode(resp) //nolint:errcheck // Test HTTP handler - encoding error in test response not critical
	}))
	defer server.Close()

	poller := NewPoller(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, true) // skipTLSVerify=true for tests
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	resp, err := poller.Poll(ctx, server.URL+"/health")
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, resp)
}

func TestPollerPollEventuallyHealthy(t *testing.T) {
	t.Parallel()

	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts < 3 {
			// First 2 attempts fail
			w.WriteHeader(http.StatusServiceUnavailable)

			return
		}
		// Third attempt succeeds
		resp := Response{Status: cryptoutilSharedMagic.DockerServiceHealthHealthy}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp) //nolint:errcheck // Test HTTP handler - encoding error in test response not critical
	}))
	defer server.Close()

	poller := NewPoller(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, true) // skipTLSVerify=true for tests
	ctx := context.Background()

	resp, err := poller.Poll(ctx, server.URL+"/health")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, cryptoutilSharedMagic.DockerServiceHealthHealthy, resp.Status)
	require.Equal(t, 3, attempts)
}

// TestNewPoller_WithTLSVerification tests NewPoller with TLS verification enabled.
func TestNewPoller_WithTLSVerification(t *testing.T) {
	t.Parallel()

	// Test skipTLSVerify=false (production setting).
	poller := NewPoller(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second, 3, false)
	require.NotNil(t, poller)
	require.NotNil(t, poller.client)
	require.Equal(t, 3*defaultInitialInterval, poller.timeout)
	require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second, poller.client.Timeout)
	require.Equal(t, defaultInitialInterval, poller.interval)
}

// TestNewPoller_WithSkipTLSVerify tests NewPoller with TLS verification disabled.
func TestNewPoller_WithSkipTLSVerify(t *testing.T) {
	t.Parallel()

	// Test skipTLSVerify=true (development/testing setting).
	poller := NewPoller(cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Second, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, true)
	require.NotNil(t, poller)
	require.NotNil(t, poller.client)
	require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*defaultInitialInterval, poller.timeout)
	require.Equal(t, cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Second, poller.client.Timeout)
	require.Equal(t, defaultInitialInterval, poller.interval)
}
