// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package apis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRateLimiter_Allow_UnderLimit(t *testing.T) {
	t.Parallel()

	// 10 requests/min, burst 5.
	rl := NewRateLimiter(10, 5)
	defer rl.Stop()

	ipAddress := "192.168.1.1"

	// First 5 requests should succeed (burst size).
	for i := 0; i < 5; i++ {
		allowed := rl.Allow(ipAddress)
		require.True(t, allowed, "Request %d should be allowed", i+1)
	}
}

func TestRateLimiter_Allow_ExceedsLimit(t *testing.T) {
	t.Parallel()

	// 10 requests/min, burst 5.
	rl := NewRateLimiter(10, 5)
	defer rl.Stop()

	ipAddress := "192.168.1.2"

	// First 5 requests succeed (burst).
	for i := 0; i < 5; i++ {
		allowed := rl.Allow(ipAddress)
		require.True(t, allowed, "Request %d should be allowed (within burst)", i+1)
	}

	// 6th request should be rate limited.
	allowed := rl.Allow(ipAddress)
	require.False(t, allowed, "Request 6 should be rate limited")
}

func TestRateLimiter_Allow_TokenRefill(t *testing.T) {
	t.Parallel()

	// 60 requests/min, burst 2 (1 token refilled per second).
	rl := NewRateLimiter(60, 2)
	defer rl.Stop()

	ipAddress := "192.168.1.3"

	// Use all tokens.
	require.True(t, rl.Allow(ipAddress))
	require.True(t, rl.Allow(ipAddress))
	require.False(t, rl.Allow(ipAddress), "Should be rate limited after using all tokens")

	// Wait for 1 second (1 token should be added).
	time.Sleep(1100 * time.Millisecond)

	// Next request should succeed (token refilled).
	allowed := rl.Allow(ipAddress)
	require.True(t, allowed, "Request should succeed after token refill")
}

func TestRateLimiter_Allow_PerIPIsolation(t *testing.T) {
	t.Parallel()

	// 10 requests/min, burst 3.
	rl := NewRateLimiter(10, 3)
	defer rl.Stop()

	ip1 := "192.168.1.4"
	ip2 := "192.168.1.5"

	// Use all tokens for IP1.
	require.True(t, rl.Allow(ip1))
	require.True(t, rl.Allow(ip1))
	require.True(t, rl.Allow(ip1))
	require.False(t, rl.Allow(ip1), "IP1 should be rate limited")

	// IP2 should have independent bucket.
	require.True(t, rl.Allow(ip2), "IP2 should be allowed (independent bucket)")
	require.True(t, rl.Allow(ip2), "IP2 should be allowed")
	require.True(t, rl.Allow(ip2), "IP2 should be allowed")
	require.False(t, rl.Allow(ip2), "IP2 should be rate limited after using all tokens")
}

func TestRateLimiter_Cleanup(t *testing.T) {
	t.Parallel()

	rl := NewRateLimiter(10, 5)
	defer rl.Stop()

	ipAddress := "192.168.1.6"

	// Create bucket.
	require.True(t, rl.Allow(ipAddress))

	// Verify bucket exists.
	rl.mu.RLock()
	_, exists := rl.buckets[ipAddress]
	rl.mu.RUnlock()
	require.True(t, exists, "Bucket should exist after request")

	// Manually set lastRefillTime to old time.
	rl.mu.Lock()
	rl.buckets[ipAddress].lastRefillTime = time.Now().Add(-15 * time.Minute)
	rl.mu.Unlock()

	// Trigger cleanup.
	rl.cleanup()

	// Verify bucket removed.
	rl.mu.RLock()
	_, exists = rl.buckets[ipAddress]
	rl.mu.RUnlock()
	require.False(t, exists, "Bucket should be removed after cleanup")
}
