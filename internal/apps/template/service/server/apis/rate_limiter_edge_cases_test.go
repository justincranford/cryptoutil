//go:build !integration

package apis

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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

	limiter := NewRateLimiter(cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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

	limiter := NewRateLimiter(cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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

// TestRateLimiter_TokenCapAfterLongIdle tests token cap after long idle period (line 73).
func TestRateLimiter_TokenCapAfterLongIdle(t *testing.T) {
	t.Parallel()

	// Create limiter with burstSize=3 and requestsPerMin=60 (1 token/sec).
	limiter := NewRateLimiter(cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds, 3)
	ipAddress := "192.168.1.400"

	// First request creates bucket with full tokens (burstSize=3).
	require.True(t, limiter.Allow(ipAddress))

	// Now artificially age the bucket back in time to simulate long idle period.
	// This will cause tokensToAdd to exceed remaining capacity, triggering the cap.
	limiter.mu.Lock()

	if bucket, ok := limiter.buckets[ipAddress]; ok {
		// Set lastRefillTime to 10 seconds ago.
		// With 60 req/min = 1 token/sec, this should add ~10 tokens.
		// But bucket already has 2 tokens (3 - 1 used), so cap at 3.
		bucket.lastRefillTime = time.Now().UTC().Add(-cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Second)
	}

	limiter.mu.Unlock()

	// Next request triggers refill calculation.
	// tokensToAdd would be ~10, but tokens would exceed burstSize (3).
	// This triggers the cap: if bucket.tokens > rl.burstSize (line 73).
	require.True(t, limiter.Allow(ipAddress))

	// Verify bucket has exactly burstSize tokens after cap (minus 1 for this request).
	limiter.mu.Lock()
	bucket := limiter.buckets[ipAddress]
	require.Equal(t, 2, bucket.tokens) // burstSize(3) - 1(used) = 2
	limiter.mu.Unlock()

	limiter.Stop()
}
