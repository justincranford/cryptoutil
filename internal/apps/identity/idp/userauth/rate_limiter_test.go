// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
	timestamp := time.Now().UTC()

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
	now := time.Now().UTC()

	// Record 3 attempts: 2 within 1 hour, 1 older.
	err = store.RecordAttempt(ctx, key, now.Add(-cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Minute))
	require.NoError(t, err)

	err = store.RecordAttempt(ctx, key, now.Add(-45*time.Minute))
	require.NoError(t, err)

	err = store.RecordAttempt(ctx, key, now.Add(-cryptoutilSharedMagic.StrictCertificateMaxAgeDays*time.Minute))
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
	now := time.Now().UTC()

	// Record attempts at different times.
	err = store.RecordAttempt(ctx, key, now.Add(-3*time.Hour))
	require.NoError(t, err)

	err = store.RecordAttempt(ctx, key, now.Add(-cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Minute))
	require.NoError(t, err)

	// Before cleanup: 2 attempts.
	count, err := store.CountAttempts(ctx, key, cryptoutilSharedMagic.HoursPerDay*time.Hour)
	require.NoError(t, err)
	require.Equal(t, 2, count)

	// Cleanup records older than 2 hours.
	err = store.CleanupExpired(ctx, 2*time.Hour)
	require.NoError(t, err)

	// After cleanup: only 1 recent attempt remains.
	count, err = store.CountAttempts(ctx, key, cryptoutilSharedMagic.HoursPerDay*time.Hour)
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

	limiter, err := NewPerUserRateLimiter(store, time.Hour, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, noop.NewMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()
	userID := googleUuid.New()

	// 5 concurrent goroutines each make 2 attempts (10 total).
	done := make(chan bool, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)

	for range cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries {
		go func() {
			defer func() { done <- true }()

			for range 2 {
				err := limiter.CheckLimit(ctx, userID)
				if err == nil {
					_ = limiter.RecordAttempt(ctx, userID) //nolint:errcheck // Concurrent test - error ignored to test race conditions
				}
			}
		}()
	}

	// Wait for all goroutines.
	for range cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries {
		<-done
	}

	// Total count should be 10 (at rate limit).
	count, err := store.CountAttempts(ctx, userID.String(), time.Hour)
	require.NoError(t, err)
	require.LessOrEqual(t, count, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, "Total attempts should not exceed rate limit")

	// Next attempt should fail.
	err = limiter.CheckLimit(ctx, userID)
	require.Error(t, err)
}

// TestPerUserRateLimiterWindowExpiration tests rate limit window expiration.
func TestPerUserRateLimiterWindowExpiration(t *testing.T) {
	store, err := NewDatabaseRateLimitStore(noop.NewMeterProvider())
	require.NoError(t, err)

	// Use very short window for test: 100ms.
	limiter, err := NewPerUserRateLimiter(store, cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond, 2, noop.NewMeterProvider())
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

	limiter, err := NewPerUserRateLimiter(store, time.Hour, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, otel.GetMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()
	userID := googleUuid.New()

	// Record old attempt.
	err = store.RecordAttempt(ctx, userID.String(), time.Now().UTC().Add(-3*time.Hour))
	require.NoError(t, err)

	// Before cleanup: 1 attempt.
	count, err := store.CountAttempts(ctx, userID.String(), cryptoutilSharedMagic.HoursPerDay*time.Hour)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	// Cleanup expired records.
	err = limiter.Cleanup(ctx)
	require.NoError(t, err)

	// After cleanup: 0 attempts (old attempt removed).
	count, err = store.CountAttempts(ctx, userID.String(), cryptoutilSharedMagic.HoursPerDay*time.Hour)
	require.NoError(t, err)
	require.Equal(t, 0, count, "Expired attempts should be cleaned up")
}

// TestPerIPRateLimiterCheckLimit tests IP-based rate limit enforcement.
func TestPerIPRateLimiterCheckLimit(t *testing.T) {
	t.Parallel()

	store, err := NewDatabaseRateLimitStore(noop.NewMeterProvider())
	require.NoError(t, err)

	limiter, err := NewPerIPRateLimiter(store, time.Hour, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, noop.NewMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()
	ipAddress := "192.168.1.100"

	// First 5 attempts should succeed.
	for range cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries {
		err = limiter.CheckLimit(ctx, ipAddress)
		require.NoError(t, err, "First 5 attempts should pass")

		err = limiter.RecordAttempt(ctx, ipAddress)
		require.NoError(t, err)
	}

	// 6th attempt should fail (rate limit exceeded).
	err = limiter.CheckLimit(ctx, ipAddress)
	require.Error(t, err)
	require.Contains(t, err.Error(), "rate limit exceeded")
	require.Contains(t, err.Error(), ipAddress)
}

// TestPerIPRateLimiterEmptyIP tests validation of empty IP address.
func TestPerIPRateLimiterEmptyIP(t *testing.T) {
	t.Parallel()

	store, err := NewDatabaseRateLimitStore(noop.NewMeterProvider())
	require.NoError(t, err)

	limiter, err := NewPerIPRateLimiter(store, time.Hour, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, noop.NewMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()

	// Empty IP should fail validation.
	err = limiter.CheckLimit(ctx, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "IP address cannot be empty")

	err = limiter.RecordAttempt(ctx, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "IP address cannot be empty")
}

// TestPerIPRateLimiterConcurrent tests concurrent IP rate limiting.
func TestPerIPRateLimiterConcurrent(t *testing.T) {
	t.Parallel()

	store, err := NewDatabaseRateLimitStore(noop.NewMeterProvider())
	require.NoError(t, err)

	limiter, err := NewPerIPRateLimiter(store, time.Hour, cryptoutilSharedMagic.MaxErrorDisplay, noop.NewMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()
	ipAddress := "10.0.0.50"

	// 10 concurrent goroutines each make 2 attempts (20 total).
	done := make(chan bool, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)

	for range cryptoutilSharedMagic.JoseJADefaultMaxMaterials {
		go func() {
			defer func() { done <- true }()

			for range 2 {
				err := limiter.CheckLimit(ctx, ipAddress)
				if err == nil {
					_ = limiter.RecordAttempt(ctx, ipAddress) //nolint:errcheck // Concurrent test - error ignored to test race conditions
				}
			}
		}()
	}

	// Wait for all goroutines.
	for range cryptoutilSharedMagic.JoseJADefaultMaxMaterials {
		<-done
	}

	// Total count should be 20 (at rate limit).
	count, err := store.CountAttempts(ctx, ipAddress, time.Hour)
	require.NoError(t, err)
	require.LessOrEqual(t, count, cryptoutilSharedMagic.MaxErrorDisplay, "Total attempts should not exceed rate limit")

	// Next attempt should fail.
	err = limiter.CheckLimit(ctx, ipAddress)
	require.Error(t, err)
}

// TestExtractIPFromContext tests IP extraction from context.
func TestExtractIPFromContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		contextValues map[any]any
		expectedIP    string
		expectError   bool
		errorContains string
	}{
		{
			name: "X-Forwarded-For single IP",
			contextValues: map[any]any{
				contextKeyXForwardedFor: "203.0.113.42",
			},
			expectedIP:  "203.0.113.42",
			expectError: false,
		},
		{
			name: "X-Forwarded-For multiple IPs",
			contextValues: map[any]any{
				contextKeyXForwardedFor: "203.0.113.42, 198.51.100.17, 192.0.2.1",
			},
			expectedIP:  "203.0.113.42",
			expectError: false,
		},
		{
			name: "RemoteAddr with port",
			contextValues: map[any]any{
				contextKeyRemoteAddr: "192.168.1.100:54321",
			},
			expectedIP:  "192.168.1.100",
			expectError: false,
		},
		{
			name: "RemoteAddr without port",
			contextValues: map[any]any{
				contextKeyRemoteAddr: "10.0.0.50",
			},
			expectedIP:  "10.0.0.50",
			expectError: false,
		},
		{
			name:          "No IP in context",
			contextValues: map[any]any{},
			expectedIP:    "",
			expectError:   true,
			errorContains: "unable to extract IP address",
		},
		{
			name: "X-Forwarded-For takes precedence over RemoteAddr",
			contextValues: map[any]any{
				contextKeyXForwardedFor: "203.0.113.42",
				contextKeyRemoteAddr:    "192.168.1.100:54321",
			},
			expectedIP:  "203.0.113.42",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			for key, value := range tc.contextValues {
				ctx = context.WithValue(ctx, key, value)
			}

			ip, err := ExtractIPFromContext(ctx)

			if tc.expectError {
				require.Error(t, err)

				if tc.errorContains != "" {
					require.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedIP, ip)
			}
		})
	}
}

// TestPerIPRateLimiterCleanup tests the Cleanup function.
func TestPerIPRateLimiterCleanup(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	meterProvider := otel.GetMeterProvider()

	store, err := NewDatabaseRateLimitStore(meterProvider)
	require.NoError(t, err, "NewDatabaseRateLimitStore should succeed")

	limiter, err := NewPerIPRateLimiter(store, time.Hour, cryptoutilSharedMagic.JoseJAMaxMaterials, meterProvider)
	require.NoError(t, err, "NewPerIPRateLimiter should succeed")

	// Record some test data.
	testIP := "192.168.1.1"
	now := time.Now().UTC()

	// Record an old attempt (older than window).
	err = store.RecordAttempt(ctx, testIP, now.Add(-3*time.Hour))
	require.NoError(t, err, "RecordAttempt should succeed")

	// Record a recent attempt (within window).
	err = store.RecordAttempt(ctx, testIP, now.Add(-cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Minute))
	require.NoError(t, err, "RecordAttempt should succeed")

	// Cleanup expired records.
	err = limiter.Cleanup(ctx)
	require.NoError(t, err, "Cleanup should succeed")

	// After cleanup, only recent attempt should remain.
	count, err := store.CountAttempts(ctx, testIP, cryptoutilSharedMagic.HoursPerDay*time.Hour)
	require.NoError(t, err, "CountAttempts should succeed")
	require.Equal(t, 1, count, "Only recent attempt should remain after cleanup")
}
