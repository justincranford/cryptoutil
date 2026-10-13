// Copyright (c) 2025-2026 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !integration

package apis

import (
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestRateLimiter_CleanupTickerFires covers cleanupLoop with an injected interval.
func TestRateLimiter_CleanupTickerFires(t *testing.T) {
	t.Parallel()

	staleBucketAge := (cryptoutilSharedMagic.RateLimitStaleThresholdMinutes + cryptoutilSharedMagic.RateLimitCleanupIntervalMinutes) * time.Minute
	freshBucketAge := cryptoutilSharedMagic.RateLimitCleanupIntervalMinutes * time.Minute

	tests := []struct {
		name             string
		targetBucketAge  time.Duration
		wantTargetExists bool
	}{
		{name: "stale target removed", targetBucketAge: staleBucketAge, wantTargetExists: false},
		{name: "fresh target retained", targetBucketAge: freshBucketAge, wantTargetExists: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			limiter := newRateLimiter(cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, time.Millisecond)
			defer limiter.Stop()

			targetIP := "10.0.0.99"
			staleSentinelIP := "10.0.0.100"

			require.True(t, limiter.Allow(targetIP))
			require.True(t, limiter.Allow(staleSentinelIP))

			limiter.mu.Lock()
			limiter.buckets[targetIP].lastRefillTime = time.Now().UTC().Add(-tc.targetBucketAge)
			limiter.buckets[staleSentinelIP].lastRefillTime = time.Now().UTC().Add(-staleBucketAge)
			limiter.mu.Unlock()

			require.Eventually(t, func() bool {
				limiter.mu.RLock()
				_, exists := limiter.buckets[staleSentinelIP]
				limiter.mu.RUnlock()

				return !exists
			}, cryptoutilSharedMagic.ClientShutdownRequestTimeout, time.Millisecond)

			limiter.mu.RLock()
			_, targetExists := limiter.buckets[targetIP]
			limiter.mu.RUnlock()

			require.Equal(t, tc.wantTargetExists, targetExists)
		})
	}
}
