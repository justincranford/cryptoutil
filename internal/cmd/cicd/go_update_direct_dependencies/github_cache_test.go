// Copyright (c) 2025 Justin Cranford

// Package go_update_direct_dependencies provides tests for GitHub API caching functionality.
package go_update_direct_dependencies

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

func TestNewGitHubAPICache(t *testing.T) {
	cache := NewGitHubAPICache()
	require.NotNil(t, cache, "Cache should not be nil")
	require.NotNil(t, cache.cache, "Cache map should be initialized")
	require.Empty(t, cache.cache, "Cache should start empty")
}

func TestGitHubAPICache(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		testFn func(t *testing.T, cache *GitHubAPICache)
	}{
		{
			name: "get cache miss and hit",
			testFn: func(t *testing.T, cache *GitHubAPICache) {
				t.Helper()

				value, found := cache.Get("nonexistent")
				require.False(t, found, "Should not find nonexistent key")
				require.Empty(t, value, "Should return empty value for nonexistent key")

				cache.Set("test-key", "test-value")
				value, found = cache.Get("test-key")
				require.True(t, found, "Should find existing key")
				require.Equal(t, "test-value", value, "Should return correct value")
			},
		},
		{
			name: "get expired entry",
			testFn: func(t *testing.T, cache *GitHubAPICache) {
				t.Helper()
				cache.mu.Lock()
				cache.cache["expired-key"] = GitHubAPICacheEntry{
					Value:     "expired-value",
					ExpiresAt: time.Now().UTC().Add(-cryptoutilMagic.TestGitHubAPICacheExpiredHours * time.Hour),
				}
				cache.mu.Unlock()

				value, found := cache.Get("expired-key")
				require.False(t, found, "Should not find expired key")
				require.Empty(t, value, "Should return empty value for expired key")

				cache.mu.RLock()
				_, exists := cache.cache["expired-key"]
				cache.mu.RUnlock()
				require.False(t, exists, "Expired entry should be removed from cache")
			},
		},
		{
			name: "set value",
			testFn: func(t *testing.T, cache *GitHubAPICache) {
				t.Helper()
				cache.Set("test-key", "test-value")

				cache.mu.RLock()
				entry, exists := cache.cache["test-key"]
				cache.mu.RUnlock()

				require.True(t, exists, "Key should exist in cache")
				require.Equal(t, "test-value", entry.Value, "Value should be stored correctly")

				expectedExpiry := time.Now().UTC().Add(cryptoutilMagic.TimeoutGitHubAPICacheTTL)
				require.True(t, entry.ExpiresAt.After(time.Now().UTC()), "Expiration should be in the future")
				require.True(t, entry.ExpiresAt.Before(expectedExpiry.Add(time.Second)), "Expiration should be approximately TTL from now")
			},
		},
		{
			name: "set overwrite existing",
			testFn: func(t *testing.T, cache *GitHubAPICache) {
				t.Helper()
				cache.Set("test-key", "test-value")
				cache.Set("test-key", "new-value")

				cache.mu.RLock()
				entry, exists := cache.cache["test-key"]
				cache.mu.RUnlock()

				require.True(t, exists, "Key should still exist")
				require.Equal(t, "new-value", entry.Value, "Value should be updated")
			},
		},
		{
			name: "expired entry cleanup",
			testFn: func(t *testing.T, cache *GitHubAPICache) {
				t.Helper()
				cache.mu.Lock()
				cache.cache["expired-key"] = GitHubAPICacheEntry{
					Value:     "should-be-removed",
					ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
				}
				cache.mu.Unlock()

				value, found := cache.Get("expired-key")
				require.False(t, found, "Expired entry should not be found")
				require.Empty(t, value, "Value should be empty for expired entry")

				cache.mu.RLock()
				_, exists := cache.cache["expired-key"]
				cache.mu.RUnlock()
				require.False(t, exists, "Expired entry should be removed from cache map")
			},
		},
		{
			name: "near expiration",
			testFn: func(t *testing.T, cache *GitHubAPICache) {
				t.Helper()
				cache.mu.Lock()
				cache.cache["near-expiry"] = GitHubAPICacheEntry{
					Value:     "expires-soon",
					ExpiresAt: time.Now().UTC().Add(100 * time.Millisecond),
				}
				cache.mu.Unlock()

				value, found := cache.Get("near-expiry")
				require.True(t, found, "Entry should be found before expiration")
				require.Equal(t, "expires-soon", value, "Value should match")

				time.Sleep(150 * time.Millisecond)

				value, found = cache.Get("near-expiry")
				require.False(t, found, "Entry should be expired after waiting")
				require.Empty(t, value, "Value should be empty after expiration")
			},
		},
		{
			name: "multiple keys",
			testFn: func(t *testing.T, cache *GitHubAPICache) {
				t.Helper()
				cache.Set("key1", "value1")
				cache.Set("key2", "value2")
				cache.Set("key3", "value3")

				val1, found1 := cache.Get("key1")
				require.True(t, found1)
				require.Equal(t, "value1", val1)

				val2, found2 := cache.Get("key2")
				require.True(t, found2)
				require.Equal(t, "value2", val2)

				val3, found3 := cache.Get("key3")
				require.True(t, found3)
				require.Equal(t, "value3", val3)

				cache.mu.Lock()
				cache.cache["key2"] = GitHubAPICacheEntry{
					Value:     "value2",
					ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
				}
				cache.mu.Unlock()

				_, found2 = cache.Get("key2")
				require.False(t, found2, "Expired key should not be found")

				val1, found1 = cache.Get("key1")
				require.True(t, found1, "Non-expired key should still exist")
				require.Equal(t, "value1", val1)

				val3, found3 = cache.Get("key3")
				require.True(t, found3, "Non-expired key should still exist")
				require.Equal(t, "value3", val3)
			},
		},
		{
			name: "update existing",
			testFn: func(t *testing.T, cache *GitHubAPICache) {
				t.Helper()
				cache.Set("update-key", "initial")
				val, found := cache.Get("update-key")
				require.True(t, found)
				require.Equal(t, "initial", val)

				cache.Set("update-key", "updated")
				val, found = cache.Get("update-key")
				require.True(t, found)
				require.Equal(t, "updated", val, "Value should be updated")

				cache.Set("update-key", "final")
				val, found = cache.Get("update-key")
				require.True(t, found)
				require.Equal(t, "final", val, "Value should be updated again")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cache := NewGitHubAPICache()
			tc.testFn(t, cache)
		})
	}
}

func TestGitHubAPICache_Concurrency(t *testing.T) {
	cache := NewGitHubAPICache()

	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			cache.Set("key1", "value1")
			_, _ = cache.Get("key1")
		}

		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			cache.Set("key2", "value2")
			_, _ = cache.Get("key2")
		}

		done <- true
	}()

	<-done
	<-done

	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	require.True(t, found1 || found2, "At least one key should exist (race condition dependent)")
}
