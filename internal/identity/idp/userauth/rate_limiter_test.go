// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/noop"

	googleUuid "github.com/google/uuid"
)

// TestDatabaseRateLimitStoreRecordAttempt tests recording attempts.
func TestDatabaseRateLimitStoreRecordAttempt(t *testing.T) {
	t.Parallel()

	store, err := NewDatabaseRateLimitStore(noop.NewMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()
	key := "user-123"
	timestamp := time.Now()

	err = store.RecordAttempt(ctx, key, timestamp)
	require.NoError(t, err)

	count, err := store.CountAttempts(ctx, key, time.Hour)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

// TestDatabaseRateLimitStoreCountAttempts tests counting attempts within window.
func TestDatabaseRateLimitStoreCountAttempts(t *testing.T) {
	t.Parallel()

	store, err := NewDatabaseRateLimitStore(noop.NewMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()
	key := "user-456"
	now := time.Now()

	// Record 3 attempts: 2 within 1 hour, 1 older.
	err = store.RecordAttempt(ctx, key, now.Add(-30*time.Minute))
	require.NoError(t, err)

	err = store.RecordAttempt(ctx, key, now.Add(-45*time.Minute))
	require.NoError(t, err)

	err = store.RecordAttempt(ctx, key, now.Add(-90*time.Minute))
	require.NoError(t, err)

	// Count within 1 hour window.
	count, err := store.CountAttempts(ctx, key, time.Hour)
	require.NoError(t, err)
	require.Equal(t, 2, count, "Should count only attempts within 1 hour")

	// Count within 2 hour window.
	count, err = store.CountAttempts(ctx, key, 2*time.Hour)
	require.NoError(t, err)
	require.Equal(t, 3, count, "Should count all attempts within 2 hours")
}

// TestDatabaseRateLimitStoreCleanupExpired tests cleanup of expired records.
func TestDatabaseRateLimitStoreCleanupExpired(t *testing.T) {
	t.Parallel()

	store, err := NewDatabaseRateLimitStore(noop.NewMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()
	key := "user-789"
	now := time.Now()

	// Record attempts at different times.
	err = store.RecordAttempt(ctx, key, now.Add(-3*time.Hour))
	require.NoError(t, err)

	err = store.RecordAttempt(ctx, key, now.Add(-30*time.Minute))
	require.NoError(t, err)

	// Before cleanup: 2 attempts.
	count, err := store.CountAttempts(ctx, key, 24*time.Hour)
	require.NoError(t, err)
	require.Equal(t, 2, count)

	// Cleanup records older than 2 hours.
	err = store.CleanupExpired(ctx, 2*time.Hour)
	require.NoError(t, err)

	// After cleanup: only 1 recent attempt remains.
	count, err = store.CountAttempts(ctx, key, 24*time.Hour)
	require.NoError(t, err)
	require.Equal(t, 1, count, "Old attempts should be cleaned up")
}

// TestPerUserRateLimiterCheckLimit tests rate limit enforcement.
func TestPerUserRateLimiterCheckLimit(t *testing.T) {
	t.Parallel()

	store, err := NewDatabaseRateLimitStore(noop.NewMeterProvider())
	require.NoError(t, err)

	limiter, err := NewPerUserRateLimiter(store, time.Hour, 3, noop.NewMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()
	userID := googleUuid.New()

	// First 3 attempts should succeed.
	for range 3 {
		err = limiter.CheckLimit(ctx, userID)
		require.NoError(t, err, "First 3 attempts should pass")

		err = limiter.RecordAttempt(ctx, userID)
		require.NoError(t, err)
	}

	// 4th attempt should fail (rate limit exceeded).
	err = limiter.CheckLimit(ctx, userID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "rate limit exceeded")
}

// TestPerUserRateLimiterConcurrent tests concurrent rate limiting.
func TestPerUserRateLimiterConcurrent(t *testing.T) {
	t.Parallel()

	store, err := NewDatabaseRateLimitStore(noop.NewMeterProvider())
	require.NoError(t, err)

	limiter, err := NewPerUserRateLimiter(store, time.Hour, 10, noop.NewMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()
	userID := googleUuid.New()

	// 5 concurrent goroutines each make 2 attempts (10 total).
	done := make(chan bool, 5)

	for range 5 {
		go func() {
			defer func() { done <- true }()

			for range 2 {
				err := limiter.CheckLimit(ctx, userID)
				if err == nil {
					_ = limiter.RecordAttempt(ctx, userID)
				}
			}
		}()
	}

	// Wait for all goroutines.
	for range 5 {
		<-done
	}

	// Total count should be 10 (at rate limit).
	count, err := store.CountAttempts(ctx, userID.String(), time.Hour)
	require.NoError(t, err)
	require.LessOrEqual(t, count, 10, "Total attempts should not exceed rate limit")

	// Next attempt should fail.
	err = limiter.CheckLimit(ctx, userID)
	require.Error(t, err)
}

// TestPerUserRateLimiterWindowExpiration tests rate limit window expiration.
func TestPerUserRateLimiterWindowExpiration(t *testing.T) {
	store, err := NewDatabaseRateLimitStore(noop.NewMeterProvider())
	require.NoError(t, err)

	// Use very short window for test: 100ms.
	limiter, err := NewPerUserRateLimiter(store, 100*time.Millisecond, 2, noop.NewMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()
	userID := googleUuid.New()

	// Make 2 attempts (hit rate limit).
	for range 2 {
		err = limiter.CheckLimit(ctx, userID)
		require.NoError(t, err)

		err = limiter.RecordAttempt(ctx, userID)
		require.NoError(t, err)
	}

	// 3rd attempt should fail.
	err = limiter.CheckLimit(ctx, userID)
	require.Error(t, err)

	// Wait for window to expire.
	time.Sleep(150 * time.Millisecond)

	// After window expiration, attempts should succeed again.
	err = limiter.CheckLimit(ctx, userID)
	require.NoError(t, err, "Rate limit should reset after window expiration")
}

// TestPerUserRateLimiterCleanup tests cleanup of expired rate limit records.
func TestPerUserRateLimiterCleanup(t *testing.T) {
	t.Parallel()

	store, err := NewDatabaseRateLimitStore(noop.NewMeterProvider())
	require.NoError(t, err)

	limiter, err := NewPerUserRateLimiter(store, time.Hour, 5, otel.GetMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()
	userID := googleUuid.New()

	// Record old attempt.
	err = store.RecordAttempt(ctx, userID.String(), time.Now().Add(-3*time.Hour))
	require.NoError(t, err)

	// Before cleanup: 1 attempt.
	count, err := store.CountAttempts(ctx, userID.String(), 24*time.Hour)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	// Cleanup expired records.
	err = limiter.Cleanup(ctx)
	require.NoError(t, err)

	// After cleanup: 0 attempts (old attempt removed).
	count, err = store.CountAttempts(ctx, userID.String(), 24*time.Hour)
	require.NoError(t, err)
	require.Equal(t, 0, count, "Expired attempts should be cleaned up")
}
