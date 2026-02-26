// Copyright (c) 2025 Justin Cranford

package ratelimit_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	cryptoutilIdentityRatelimit "cryptoutil/internal/apps/identity/ratelimit"

	"github.com/stretchr/testify/require"
)

func TestRateLimiter_Allow(t *testing.T) {
	t.Parallel()

	rl := cryptoutilIdentityRatelimit.NewRateLimiter(3, cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond)

	// First 3 requests should succeed.
	for i := 0; i < 3; i++ {
		err := rl.Allow("user1")
		require.NoError(t, err, "Request %d should be allowed", i+1)
	}

	// 4th request should fail (rate limit exceeded).
	err := rl.Allow("user1")
	require.Error(t, err, "4th request should exceed rate limit")
	require.Contains(t, err.Error(), "rate limit exceeded")
}

func TestRateLimiter_Allow_DifferentKeys(t *testing.T) {
	t.Parallel()

	rl := cryptoutilIdentityRatelimit.NewRateLimiter(2, cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond)

	// User1: 2 requests (at limit).
	err := rl.Allow("user1")
	require.NoError(t, err)
	err = rl.Allow("user1")
	require.NoError(t, err)

	// User2: 2 requests (at limit).
	err = rl.Allow("user2")
	require.NoError(t, err)
	err = rl.Allow("user2")
	require.NoError(t, err)

	// User1: 3rd request should fail.
	err = rl.Allow("user1")
	require.Error(t, err)

	// User2: 3rd request should fail.
	err = rl.Allow("user2")
	require.Error(t, err)
}

func TestRateLimiter_WindowExpiration(t *testing.T) {
	t.Parallel()

	rl := cryptoutilIdentityRatelimit.NewRateLimiter(2, cryptoutilSharedMagic.IMMaxUsernameLength*time.Millisecond)

	// Make 2 requests (at limit).
	err := rl.Allow("user1")
	require.NoError(t, err)
	err = rl.Allow("user1")
	require.NoError(t, err)

	// 3rd request should fail.
	err = rl.Allow("user1")
	require.Error(t, err)

	// Wait for window to expire.
	time.Sleep(cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds * time.Millisecond)

	// After window expiration, request should succeed.
	err = rl.Allow("user1")
	require.NoError(t, err, "Request should succeed after window expiration")
}

func TestRateLimiter_Reset(t *testing.T) {
	t.Parallel()

	rl := cryptoutilIdentityRatelimit.NewRateLimiter(2, cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond)

	// Make 2 requests (at limit).
	err := rl.Allow("user1")
	require.NoError(t, err)
	err = rl.Allow("user1")
	require.NoError(t, err)

	// 3rd request should fail.
	err = rl.Allow("user1")
	require.Error(t, err)

	// Reset rate limit for user1.
	rl.Reset("user1")

	// After reset, request should succeed.
	err = rl.Allow("user1")
	require.NoError(t, err, "Request should succeed after reset")
}

func TestRateLimiter_GetCount(t *testing.T) {
	t.Parallel()

	rl := cryptoutilIdentityRatelimit.NewRateLimiter(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond)

	// No requests yet.
	count := rl.GetCount("user1")
	require.Equal(t, 0, count)

	// Make 3 requests.
	for i := 0; i < 3; i++ {
		err := rl.Allow("user1")
		require.NoError(t, err)
	}

	// Count should be 3.
	count = rl.GetCount("user1")
	require.Equal(t, 3, count)

	// Wait for window to expire.
	time.Sleep(110 * time.Millisecond)

	// Count should be 0 after expiration.
	count = rl.GetCount("user1")
	require.Equal(t, 0, count)
}
