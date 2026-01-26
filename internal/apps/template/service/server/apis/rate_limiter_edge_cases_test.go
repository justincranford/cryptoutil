//go:build !integration

package apis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestRateLimiter_ExhaustTokens tests false return path (line 85).
func TestRateLimiter_ExhaustTokens(t *testing.T) {
	t.Parallel()

	limiter := NewRateLimiter(2, 1) // 2 requests/min, burst 1
	ipAddress := "192.168.1.100"

	// First request: should succeed (uses burst token).
	allowed := limiter.Allow(ipAddress)
	require.True(t, allowed, "First request should be allowed")

	// Second request: should fail (no tokens left, refill not happened yet).
	allowed = limiter.Allow(ipAddress)
	require.False(t, allowed, "Second request should be rate limited")
}

// TestRateLimiter_CleanupStale tests cleanup of stale buckets.
func TestRateLimiter_CleanupStale(t *testing.T) {
	t.Parallel()

	limiter := NewRateLimiter(60, 10)
	ipAddress := "192.168.1.300"

	// Create bucket.
	limiter.Allow(ipAddress)

	// Manually trigger cleanup immediately (bucket not stale yet).
	limiter.cleanup()

	// Bucket should still exist (not stale).
	limiter.mu.Lock()
	_, exists := limiter.buckets[ipAddress]
	limiter.mu.Unlock()
	require.True(t, exists, "Bucket should not be cleaned up (not stale)")

	// Artificially age the bucket by modifying lastRefillTime.
	limiter.mu.Lock()

	if bucket, ok := limiter.buckets[ipAddress]; ok {
		bucket.lastRefillTime = time.Now().UTC().Add(-15 * time.Minute) // > 10 min threshold
	}

	limiter.mu.Unlock()

	// Trigger cleanup again.
	limiter.cleanup()

	// Bucket should now be removed.
	limiter.mu.Lock()
	_, exists = limiter.buckets[ipAddress]
	limiter.mu.Unlock()
	require.False(t, exists, "Stale bucket should be cleaned up")

	limiter.Stop()
}

// TestRateLimiter_StopCleanupLoop tests cleanup goroutine termination (line 95).
func TestRateLimiter_StopCleanupLoop(t *testing.T) {
	t.Parallel()

	limiter := NewRateLimiter(60, 10)
	ipAddress := "192.168.1.200"

	// Generate some traffic.
	limiter.Allow(ipAddress)

	// Stop the limiter (triggers stopCleanup channel case).
	limiter.Stop()

	// Give goroutine time to exit via stopCleanup channel (line 95).
	time.Sleep(200 * time.Millisecond)

	// Test passes if no deadlock occurs.
	// This verifies cleanupLoop received stopCleanup signal and exited.
}
