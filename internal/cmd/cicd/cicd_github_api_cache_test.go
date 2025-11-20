// Package cicd provides tests for GitHub API caching functionality.
package cicd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/cmd/cicd/common"
	cryptoutilMagic "cryptoutil/internal/common/magic"
)

func TestNewGitHubAPICache(t *testing.T) {
	cache := NewGitHubAPICache()
	require.NotNil(t, cache, "Cache should not be nil")
	require.NotNil(t, cache.cache, "Cache map should be initialized")
	require.Empty(t, cache.cache, "Cache should start empty")
}

func TestGitHubAPICache_Get(t *testing.T) {
	cache := NewGitHubAPICache()

	// Test cache miss
	value, found := cache.Get("nonexistent")
	require.False(t, found, "Should not find nonexistent key")
	require.Empty(t, value, "Should return empty value for nonexistent key")

	// Test cache hit
	cache.Set("test-key", "test-value")
	value, found = cache.Get("test-key")
	require.True(t, found, "Should find existing key")
	require.Equal(t, "test-value", value, "Should return correct value")

	// Test expired entry (simulate by setting expired time)
	cache.mu.Lock()
	cache.cache["expired-key"] = GitHubAPICacheEntry{
		Value:     "expired-value",
		ExpiresAt: time.Now().UTC().Add(-cryptoutilMagic.TestGitHubAPICacheExpiredHours * time.Hour), // Expired 1 hour ago
	}
	cache.mu.Unlock()

	value, found = cache.Get("expired-key")
	require.False(t, found, "Should not find expired key")
	require.Empty(t, value, "Should return empty value for expired key")

	// Verify expired entry was removed
	cache.mu.RLock()
	_, exists := cache.cache["expired-key"]
	cache.mu.RUnlock()
	require.False(t, exists, "Expired entry should be removed from cache")
}

func TestGitHubAPICache_Set(t *testing.T) {
	cache := NewGitHubAPICache()

	// Set a value
	cache.Set("test-key", "test-value")

	// Verify it was set
	cache.mu.RLock()
	entry, exists := cache.cache["test-key"]
	cache.mu.RUnlock()

	require.True(t, exists, "Key should exist in cache")
	require.Equal(t, "test-value", entry.Value, "Value should be stored correctly")

	// Verify expiration time is set correctly (should be now + TTL)
	expectedExpiry := time.Now().UTC().Add(cryptoutilMagic.TimeoutGitHubAPICacheTTL)
	require.True(t, entry.ExpiresAt.After(time.Now().UTC()), "Expiration should be in the future")
	require.True(t, entry.ExpiresAt.Before(expectedExpiry.Add(time.Second)), "Expiration should be approximately TTL from now")

	// Test overwriting existing value
	cache.Set("test-key", "new-value")
	cache.mu.RLock()
	entry, exists = cache.cache["test-key"]
	cache.mu.RUnlock()

	require.True(t, exists, "Key should still exist")
	require.Equal(t, "new-value", entry.Value, "Value should be updated")
}

func TestGitHubAPICache_Concurrency(t *testing.T) {
	cache := NewGitHubAPICache()

	// Test concurrent access
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

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify final state
	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	require.True(t, found1 || found2, "At least one key should exist (race condition dependent)")
}
