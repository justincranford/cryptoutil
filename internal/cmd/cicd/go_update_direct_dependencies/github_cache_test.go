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

func TestGitHubAPICache_ExpiredEntryCleanup(t *testing.T) {
	t.Parallel()

	cache := NewGitHubAPICache()

	// Set expired entry directly
	cache.mu.Lock()
	cache.cache["expired-key"] = GitHubAPICacheEntry{
		Value:     "should-be-removed",
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
	}
	cache.mu.Unlock()

	// Get should remove expired entry
	value, found := cache.Get("expired-key")
	require.False(t, found, "Expired entry should not be found")
	require.Empty(t, value, "Value should be empty for expired entry")

	// Verify cleanup happened (entry removed from map during Get)
	cache.mu.RLock()
	_, exists := cache.cache["expired-key"]
	cache.mu.RUnlock()
	require.False(t, exists, "Expired entry should be removed from cache map")
}

func TestGitHubAPICache_NearExpiration(t *testing.T) {
	t.Parallel()

	cache := NewGitHubAPICache()

	// Set entry that expires in 1 second
	cache.mu.Lock()
	cache.cache["near-expiry"] = GitHubAPICacheEntry{
		Value:     "expires-soon",
		ExpiresAt: time.Now().UTC().Add(100 * time.Millisecond),
	}
	cache.mu.Unlock()

	// Should be valid immediately
	value, found := cache.Get("near-expiry")
	require.True(t, found, "Entry should be found before expiration")
	require.Equal(t, "expires-soon", value, "Value should match")

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	value, found = cache.Get("near-expiry")
	require.False(t, found, "Entry should be expired after waiting")
	require.Empty(t, value, "Value should be empty after expiration")
}

func TestGitHubAPICache_MultipleKeys(t *testing.T) {
	t.Parallel()

	cache := NewGitHubAPICache()

	// Set multiple keys
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Verify all keys exist
	val1, found1 := cache.Get("key1")
	require.True(t, found1)
	require.Equal(t, "value1", val1)

	val2, found2 := cache.Get("key2")
	require.True(t, found2)
	require.Equal(t, "value2", val2)

	val3, found3 := cache.Get("key3")
	require.True(t, found3)
	require.Equal(t, "value3", val3)

	// Expire one key
	cache.mu.Lock()
	cache.cache["key2"] = GitHubAPICacheEntry{
		Value:     "value2",
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
	}
	cache.mu.Unlock()

	// Verify expired key removed, others remain
	_, found2 = cache.Get("key2")
	require.False(t, found2, "Expired key should not be found")

	val1, found1 = cache.Get("key1")
	require.True(t, found1, "Non-expired key should still exist")
	require.Equal(t, "value1", val1)

	val3, found3 = cache.Get("key3")
	require.True(t, found3, "Non-expired key should still exist")
	require.Equal(t, "value3", val3)
}

func TestGitHubAPICache_UpdateExisting(t *testing.T) {
	t.Parallel()

	cache := NewGitHubAPICache()

	// Set initial value
	cache.Set("update-key", "initial")
	val, found := cache.Get("update-key")
	require.True(t, found)
	require.Equal(t, "initial", val)

	// Update value
	cache.Set("update-key", "updated")
	val, found = cache.Get("update-key")
	require.True(t, found)
	require.Equal(t, "updated", val, "Value should be updated")

	// Update again
	cache.Set("update-key", "final")
	val, found = cache.Get("update-key")
	require.True(t, found)
	require.Equal(t, "final", val, "Value should be updated again")
}
