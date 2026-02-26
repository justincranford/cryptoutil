// Copyright (c) 2025 Justin Cranford
//
//

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

func TestBatchIntrospector_NewBatchIntrospector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  IntrospectionConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: IntrospectionConfig{
				IntrospectionURL: "https://identity.example.com/introspect",
			},
			wantErr: false,
		},
		{
			name:    "missing URL",
			config:  IntrospectionConfig{},
			wantErr: true,
		},
		{
			name: "with all options",
			config: IntrospectionConfig{
				IntrospectionURL: "https://identity.example.com/introspect",
				ClientID:         "client-id",
				ClientSecret:     "client-secret",
				CacheTTL:         time.Minute,
				MaxBatchSize:     cryptoutilSharedMagic.MaxErrorDisplay,
				HTTPTimeout:      time.Second * cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			introspector, err := NewBatchIntrospector(tc.config)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, introspector)
			} else {
				require.NoError(t, err)
				require.NotNil(t, introspector)
			}
		})
	}
}

func TestBatchIntrospector_Introspect(t *testing.T) {
	t.Parallel()

	// Create mock server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		token := r.FormValue(cryptoutilSharedMagic.ParamToken)
		response := IntrospectionResult{Active: token == "valid-token"}

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))

	defer server.Close()

	introspector, err := NewBatchIntrospector(IntrospectionConfig{
		IntrospectionURL: server.URL,
		CacheTTL:         time.Minute,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test valid token.
	result, err := introspector.Introspect(ctx, "valid-token")
	require.NoError(t, err)
	require.True(t, result.Active)

	// Test invalid token.
	result, err = introspector.Introspect(ctx, "invalid-token")
	require.NoError(t, err)
	require.False(t, result.Active)

	// Test caching - second call should use cache.
	require.Equal(t, 2, introspector.CacheSize())
}

func TestBatchIntrospector_BatchIntrospect(t *testing.T) {
	t.Parallel()

	requestCount := 0

	// Create mock server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		token := r.FormValue(cryptoutilSharedMagic.ParamToken)
		response := IntrospectionResult{Active: token == "valid-token-1" || token == "valid-token-2"}

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(response)
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

	// Test batch with duplicates.
	tokens := []string{
		"valid-token-1",
		"valid-token-2",
		"invalid-token",
		"valid-token-1", // Duplicate.
	}

	results, err := introspector.BatchIntrospect(ctx, tokens)
	require.NoError(t, err)
	require.Len(t, results, 3) // Deduplicated.

	require.True(t, results["valid-token-1"].Active)
	require.True(t, results["valid-token-2"].Active)
	require.False(t, results["invalid-token"].Active)

	// Verify deduplication worked (only 3 requests).
	require.Equal(t, 3, requestCount)

	// Test caching on second call.
	requestCount = 0

	_, err = introspector.BatchIntrospect(ctx, tokens)
	require.NoError(t, err)
	require.Equal(t, 0, requestCount) // All from cache.
}

func TestBatchIntrospector_ClearCache(t *testing.T) {
	t.Parallel()

	introspector, err := NewBatchIntrospector(IntrospectionConfig{
		IntrospectionURL: "https://example.com/introspect",
	})
	require.NoError(t, err)

	// Add entries to cache.
	introspector.setCached("token1", true)
	introspector.setCached("token2", false)
	require.Equal(t, 2, introspector.CacheSize())

	// Clear cache.
	introspector.ClearCache()
	require.Equal(t, 0, introspector.CacheSize())
}

func TestBatchIntrospector_CacheExpiry(t *testing.T) {
	t.Parallel()

	introspector, err := NewBatchIntrospector(IntrospectionConfig{
		IntrospectionURL: "https://example.com/introspect",
		CacheTTL:         cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Millisecond,
	})
	require.NoError(t, err)

	// Add entry to cache.
	introspector.setCached(cryptoutilSharedMagic.ParamToken, true)

	// Entry should be cached.
	entry := introspector.getCached(cryptoutilSharedMagic.ParamToken)
	require.NotNil(t, entry)
	require.True(t, entry.active)

	// Wait for expiry.
	time.Sleep(cryptoutilSharedMagic.MaxErrorDisplay * time.Millisecond)

	// Entry should be expired.
	entry = introspector.getCached(cryptoutilSharedMagic.ParamToken)
	require.Nil(t, entry)
}

func TestDefaultIntrospectionConfig(t *testing.T) {
	t.Parallel()

	config := DefaultIntrospectionConfig()

	require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute, config.CacheTTL)
	require.Equal(t, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, config.MaxBatchSize)
	require.Equal(t, cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Second, config.HTTPTimeout)
}

func TestBatchIntrospector_DeduplicateTokens(t *testing.T) {
	t.Parallel()

	introspector, err := NewBatchIntrospector(IntrospectionConfig{
		IntrospectionURL: "https://example.com/introspect",
	})
	require.NoError(t, err)

	tests := []struct {
		name   string
		tokens []string
		want   int
	}{
		{
			name:   "no duplicates",
			tokens: []string{"a", "b", "c"},
			want:   3,
		},
		{
			name:   "with duplicates",
			tokens: []string{"a", "b", "a", "c", "b"},
			want:   3,
		},
		{
			name:   "all same",
			tokens: []string{"a", "a", "a"},
			want:   1,
		},
		{
			name:   "empty",
			tokens: []string{},
			want:   0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := introspector.deduplicateTokens(tc.tokens)
			require.Len(t, result, tc.want)
		})
	}
}
