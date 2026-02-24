// Copyright (c) 2025 Justin Cranford

package healthcheck

import (
	"context"
	json "encoding/json"
	"errors"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"

	"github.com/stretchr/testify/require"
)

func TestPoller_IntervalCapping(t *testing.T) {
	t.Parallel()

	// Server always returns unhealthy, forcing max retries.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := Response{Status: "unhealthy"}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(resp) //nolint:errcheck // Test HTTP handler.
	}))
	defer server.Close()

	// Use small intervals so poll.Until retries quickly.
	poller := &Poller{
		client:   server.Client(),
		timeout:  10 * time.Millisecond,
		interval: 1 * time.Millisecond,
	}

	resp, err := poller.Poll(context.Background(), server.URL+"/health")
	require.Error(t, err)
	require.Contains(t, err.Error(), "health check failed")
	require.Nil(t, resp)
}

func TestPoller_InvalidURL(t *testing.T) {
	t.Parallel()

	poller := &Poller{
		client:   &http.Client{Timeout: 1 * time.Second},
		timeout:  5 * time.Millisecond,
		interval: 1 * time.Millisecond,
	}

	// URL with null byte causes http.NewRequestWithContext to fail.
	resp, err := poller.Poll(context.Background(), "http://\x00invalid")
	require.Error(t, err)
	require.Contains(t, err.Error(), "health check failed")
	require.Nil(t, resp)
}

func TestPoller_ClosedServer(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	serverURL := server.URL
	server.Close() // Close immediately so client.Do fails.

	poller := &Poller{
		client:   &http.Client{Timeout: 1 * time.Second},
		timeout:  5 * time.Millisecond,
		interval: 1 * time.Millisecond,
	}

	resp, err := poller.Poll(context.Background(), serverURL+"/health")
	require.Error(t, err)
	require.Contains(t, err.Error(), "health check failed")
	require.Nil(t, resp)
}

func TestNewPoller_TLSConfigFallback(t *testing.T) {
	t.Parallel()

	// Save original and restore after test.
	originalFn := newClientConfigFn

	defer func() { newClientConfigFn = originalFn }()

	newClientConfigFn = func(_ *cryptoutilSharedCryptoTls.ClientConfigOptions) (*cryptoutilSharedCryptoTls.Config, error) {
		return nil, errors.New("simulated TLS config failure")
	}

	poller := NewPoller(5*time.Second, 3, true)
	require.NotNil(t, poller)
	require.NotNil(t, poller.client)
	require.Equal(t, 3*defaultInitialInterval, poller.timeout)
	require.Equal(t, 5*time.Second, poller.client.Timeout)
	// When TLS config fails, transport should be nil (no TLS config).
	require.Nil(t, poller.client.Transport)
}
