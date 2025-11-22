// Copyright (c) 2025 Justin Cranford

package healthcheck

import (
	"context"
	"encoding/json"
	"net/http"
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
			Status:   "healthy",
			Database: "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	poller := NewPoller(5*time.Second, 3)
	ctx := context.Background()

	resp, err := poller.Poll(ctx, server.URL+"/health")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "healthy", resp.Status)
	require.Equal(t, "ok", resp.Database)
}

func TestPollerPollUnhealthy(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response{
			Status: "unhealthy",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	poller := NewPoller(5*time.Second, 2)
	ctx := context.Background()

	resp, err := poller.Poll(ctx, server.URL+"/health")
	require.Error(t, err)
	require.Contains(t, err.Error(), "health check failed after")
	require.Nil(t, resp)
}

func TestPollerPollNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	poller := NewPoller(5*time.Second, 2)
	ctx := context.Background()

	resp, err := poller.Poll(ctx, server.URL+"/health")
	require.Error(t, err)
	require.Contains(t, err.Error(), "health check failed after")
	require.Nil(t, resp)
}

func TestPollerPollInvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json{{{"))
	}))
	defer server.Close()

	poller := NewPoller(5*time.Second, 2)
	ctx := context.Background()

	resp, err := poller.Poll(ctx, server.URL+"/health")
	require.Error(t, err)
	require.Contains(t, err.Error(), "health check failed after")
	require.Nil(t, resp)
}

func TestPollerPollContextCanceled(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		resp := Response{Status: "healthy"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	poller := NewPoller(5*time.Second, 10)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	resp, err := poller.Poll(ctx, server.URL+"/health")
	require.Error(t, err)
	require.Equal(t, context.Canceled, err)
	require.Nil(t, resp)
}

func TestPollerPollEventuallyHealthy(t *testing.T) {
	t.Parallel()

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// First 2 attempts fail
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// Third attempt succeeds
		resp := Response{Status: "healthy"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	poller := NewPoller(5*time.Second, 5)
	ctx := context.Background()

	resp, err := poller.Poll(ctx, server.URL+"/health")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "healthy", resp.Status)
	require.Equal(t, 3, attempts)
}
