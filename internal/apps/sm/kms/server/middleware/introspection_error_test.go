// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package middleware

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testValidToken = "valid-token"

// TestBatchIntrospector_IntrospectServerError tests the error path when the introspection server returns an error.
func TestBatchIntrospector_IntrospectServerError(t *testing.T) {
	t.Parallel()

	// Create mock server that returns 500.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	defer server.Close()

	introspector, err := NewBatchIntrospector(IntrospectionConfig{
		IntrospectionURL: server.URL,
		CacheTTL:         time.Minute,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Should return error because server returns 500.
	result, err := introspector.Introspect(ctx, "test-token")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "introspection returned status 500")
}

// TestBatchIntrospector_IntrospectInvalidJSON tests the error path when the server returns invalid JSON.
func TestBatchIntrospector_IntrospectInvalidJSON(t *testing.T) {
	t.Parallel()

	// Create mock server that returns invalid JSON.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not-valid-json"))
	}))

	defer server.Close()

	introspector, err := NewBatchIntrospector(IntrospectionConfig{
		IntrospectionURL: server.URL,
		CacheTTL:         time.Minute,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Should return error because response is not valid JSON.
	result, err := introspector.Introspect(ctx, "test-token")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to parse introspection response")
}

// TestBatchIntrospector_IntrospectConnectionRefused tests the error path when the server is unreachable.
func TestBatchIntrospector_IntrospectConnectionRefused(t *testing.T) {
	t.Parallel()

	introspector, err := NewBatchIntrospector(IntrospectionConfig{
		IntrospectionURL: "http://127.0.0.1:1", // Port 1 should be unreachable.
		CacheTTL:         time.Minute,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Should return error because connection is refused.
	result, err := introspector.Introspect(ctx, "test-token")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "introspection request failed")
}

// TestBatchIntrospector_IntrospectWithClientAuth tests the client authentication path.
func TestBatchIntrospector_IntrospectWithClientAuth(t *testing.T) {
	t.Parallel()

	// Create mock server that verifies basic auth.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		require.True(t, ok, "Basic auth should be present")
		require.Equal(t, cryptoutilSharedMagic.TestClientID, username)
		require.Equal(t, "test-client-secret", password)

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(IntrospectionResult{Active: true})
		require.NoError(t, err)
	}))

	defer server.Close()

	introspector, err := NewBatchIntrospector(IntrospectionConfig{
		IntrospectionURL: server.URL,
		CacheTTL:         time.Minute,
		ClientID:         cryptoutilSharedMagic.TestClientID,
		ClientSecret:     "test-client-secret",
	})
	require.NoError(t, err)

	ctx := context.Background()

	result, err := introspector.Introspect(ctx, "test-token")
	require.NoError(t, err)
	require.True(t, result.Active)
}

// TestBatchIntrospector_BatchIntrospectWithErrors tests the batch error continuation path.
func TestBatchIntrospector_BatchIntrospectWithErrors(t *testing.T) {
	t.Parallel()

	requestCount := 0

	// Create mock server that fails for specific tokens.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		token := r.FormValue(cryptoutilSharedMagic.ParamToken)

		if token == "error-token" {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(IntrospectionResult{Active: token == testValidToken})
		require.NoError(t, err)
	}))

	defer server.Close()

	introspector, err := NewBatchIntrospector(IntrospectionConfig{
		IntrospectionURL: server.URL,
		CacheTTL:         time.Minute,
		MaxBatchSize:     cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Batch with one token that causes server error.
	tokens := []string{testValidToken, "error-token", "invalid-token"}

	results, err := introspector.BatchIntrospect(ctx, tokens)
	require.NoError(t, err)
	require.Len(t, results, 3)

	// Valid token should be active.
	require.True(t, results[testValidToken].Active)

	// Error token should be marked inactive (continue on error in processBatch).
	require.False(t, results["error-token"].Active)

	// Invalid token should be inactive.
	require.False(t, results["invalid-token"].Active)

	// All 3 tokens should have been requested.
	require.Equal(t, 3, requestCount)
}

// TestBatchIntrospector_BatchIntrospectAllCached tests the early return when all tokens are cached.
func TestBatchIntrospector_BatchIntrospectAllCached(t *testing.T) {
	t.Parallel()

	requestCount := 0

	// Create mock server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(IntrospectionResult{Active: true})
		require.NoError(t, err)
	}))

	defer server.Close()

	introspector, err := NewBatchIntrospector(IntrospectionConfig{
		IntrospectionURL: server.URL,
		CacheTTL:         time.Minute,
		MaxBatchSize:     cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// First call - populates cache.
	_, err = introspector.Introspect(ctx, "token-1")
	require.NoError(t, err)
	require.Equal(t, 1, requestCount)

	// Batch with only cached tokens.
	results, err := introspector.BatchIntrospect(ctx, []string{"token-1"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.True(t, results["token-1"].Active)

	// No additional requests should have been made.
	require.Equal(t, 1, requestCount)
}

// TestBatchIntrospector_BatchIntrospectMultipleBatches tests batch processing across multiple batch sizes.
func TestBatchIntrospector_BatchIntrospectMultipleBatches(t *testing.T) {
	t.Parallel()

	requestCount := 0

	// Create mock server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(IntrospectionResult{Active: true})
		require.NoError(t, err)
	}))

	defer server.Close()

	introspector, err := NewBatchIntrospector(IntrospectionConfig{
		IntrospectionURL: server.URL,
		CacheTTL:         time.Minute,
		MaxBatchSize:     2, // Small batch size to force multiple batches.
	})
	require.NoError(t, err)

	ctx := context.Background()

	// 5 tokens with batch size 2 = 3 batches (2+2+1).
	tokens := []string{"token-1", "token-2", "token-3", "token-4", "token-5"}

	results, err := introspector.BatchIntrospect(ctx, tokens)
	require.NoError(t, err)
	require.Len(t, results, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)

	// All tokens should be active.
	for _, token := range tokens {
		require.True(t, results[token].Active)
	}

	// All 5 tokens should have been requested individually.
	require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, requestCount)
}

// TestBatchIntrospector_IntrospectCacheHit tests the Introspect cache hit path.
func TestBatchIntrospector_IntrospectCacheHit(t *testing.T) {
	t.Parallel()

	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"active": true}`))
	}))
	defer server.Close()

	introspector, introspectorErr := NewBatchIntrospector(IntrospectionConfig{
		IntrospectionURL: server.URL,
		MaxBatchSize:     cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
	})
	require.NoError(t, introspectorErr)

	ctx := context.Background()

	// First call should hit the server.
	result1, err := introspector.Introspect(ctx, "cached-token")
	require.NoError(t, err)
	require.True(t, result1.Active)
	require.Equal(t, 1, requestCount)

	// Second call should use cache (no additional server request).
	result2, err := introspector.Introspect(ctx, "cached-token")
	require.NoError(t, err)
	require.True(t, result2.Active)
	require.Equal(t, 1, requestCount) // Still 1 because of cache hit.
}
