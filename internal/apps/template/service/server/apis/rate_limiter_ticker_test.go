// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !integration

package apis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestRateLimiter_CleanupTickerFires covers the ticker case in cleanupLoop
// (rate_limiter.go:93: rl.cleanup()). The default ticker interval is 5 minutes,
// so we replace it with a 1ms ticker via direct field access (same package).
func TestRateLimiter_CleanupTickerFires(t *testing.T) {
	t.Parallel()

	limiter := NewRateLimiter(60, 10)
	ipAddress := "10.0.0.99"

	// Create a bucket so cleanup has something to find.
	limiter.Allow(ipAddress)

	// Age the bucket to be stale (> 10 min threshold).
	limiter.mu.Lock()

	if bucket, ok := limiter.buckets[ipAddress]; ok {
		bucket.lastRefillTime = time.Now().UTC().Add(-15 * time.Minute)
	}

	limiter.mu.Unlock()

	// Replace the default ticker (5-minute interval) with a 1ms ticker.
	limiter.cleanupTicker.Stop()
	limiter.cleanupTicker = time.NewTicker(time.Millisecond)

	// Wait for the ticker to fire and cleanupLoop to call cleanup().
	// The cleanup will remove our stale bucket.
	deadline := time.Now().UTC().Add(2 * time.Second)
	for time.Now().UTC().Before(deadline) {
		time.Sleep(5 * time.Millisecond)

		limiter.mu.RLock()
		_, exists := limiter.buckets[ipAddress]
		limiter.mu.RUnlock()

		if !exists {
			// Cleanup fired and removed the stale bucket — test passes.
			limiter.Stop()

			return
		}
	}

	limiter.Stop()
	require.Fail(t, "cleanup via ticker did not remove stale bucket within 2 seconds")
}
