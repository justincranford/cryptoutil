// Copyright (c) 2025-2026 Justin Cranford.
//
//

package userauth

import (
	"context"
	"errors"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const (
	rateLimitTestNameCountAttempts         = "count attempts"
	rateLimitTestNameRecordAttempt         = "record attempt"
	rateLimitTestNameCleanupExpiredRecords = "cleanup expired records"
	rateLimitTestInjectedCountFailure      = "injected count failure"
	rateLimitTestInjectedRecordFailure     = "injected record failure"
	rateLimitTestInjectedCleanupFailure    = "injected cleanup failure"
	rateLimitTestIPAddressCannotBeEmpty    = "IP address cannot be empty"
	rateLimitTestIPAddress                 = "192.0.2.10"
)

type rateLimitErrorStore struct {
	countErr   error
	recordErr  error
	cleanupErr error
}

func (s *rateLimitErrorStore) RecordAttempt(context.Context, string, time.Time) error {
	return s.recordErr
}

func (s *rateLimitErrorStore) CountAttempts(context.Context, string, time.Duration) (int, error) {
	return 0, s.countErr
}

func (s *rateLimitErrorStore) CleanupExpired(context.Context, time.Duration) error {
	return s.cleanupErr
}

type rateLimitErrorMeterProvider struct {
	metric.MeterProvider
}

func (p rateLimitErrorMeterProvider) Meter(name string, options ...metric.MeterOption) metric.Meter {
	return rateLimitErrorMeter{Meter: p.MeterProvider.Meter(name, options...)}
}

type rateLimitErrorMeter struct {
	metric.Meter
}

func (m rateLimitErrorMeter) Int64Counter(string, ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	return nil, errors.New("injected counter creation failure")
}

func TestRateLimiters_CounterCreationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		build func(metric.MeterProvider) error
	}{
		{
			name: "database store",
			build: func(provider metric.MeterProvider) error {
				_, err := NewDatabaseRateLimitStore(provider)

				return err
			},
		},
		{
			name: "per-user limiter",
			build: func(provider metric.MeterProvider) error {
				_, err := NewPerUserRateLimiter(&rateLimitErrorStore{}, time.Minute, 1, provider)

				return err
			},
		},
		{
			name: "per-IP limiter",
			build: func(provider metric.MeterProvider) error {
				_, err := NewPerIPRateLimiter(&rateLimitErrorStore{}, time.Minute, 1, provider)

				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.build(rateLimitErrorMeterProvider{MeterProvider: noop.NewMeterProvider()})
			require.ErrorContains(t, err, "injected counter creation failure")
		})
	}
}

func TestPerUserRateLimiter_StoreErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*rateLimitErrorStore)
		invoke    func(*PerUserRateLimiter, context.Context, googleUuid.UUID) error
		wantError string
	}{
		{
			name: rateLimitTestNameCountAttempts,
			configure: func(store *rateLimitErrorStore) {
				store.countErr = errors.New(rateLimitTestInjectedCountFailure)
			},
			invoke: func(limiter *PerUserRateLimiter, ctx context.Context, userID googleUuid.UUID) error {
				return limiter.CheckLimit(ctx, userID)
			},
			wantError: rateLimitTestInjectedCountFailure,
		},
		{
			name: rateLimitTestNameRecordAttempt,
			configure: func(store *rateLimitErrorStore) {
				store.recordErr = errors.New(rateLimitTestInjectedRecordFailure)
			},
			invoke: func(limiter *PerUserRateLimiter, ctx context.Context, userID googleUuid.UUID) error {
				return limiter.RecordAttempt(ctx, userID)
			},
			wantError: rateLimitTestInjectedRecordFailure,
		},
		{
			name: rateLimitTestNameCleanupExpiredRecords,
			configure: func(store *rateLimitErrorStore) {
				store.cleanupErr = errors.New(rateLimitTestInjectedCleanupFailure)
			},
			invoke: func(limiter *PerUserRateLimiter, ctx context.Context, userID googleUuid.UUID) error {
				return limiter.Cleanup(ctx)
			},
			wantError: rateLimitTestInjectedCleanupFailure,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := &rateLimitErrorStore{}
			tc.configure(store)

			limiter, err := NewPerUserRateLimiter(store, time.Minute, 1, noop.NewMeterProvider())
			require.NoError(t, err)

			userID, err := googleUuid.NewV7()
			require.NoError(t, err)

			err = tc.invoke(limiter, context.Background(), userID)
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestPerIPRateLimiter_ValidationAndStoreErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*rateLimitErrorStore)
		invoke    func(*PerIPRateLimiter, context.Context, string) error
		wantError string
	}{
		{
			name: "empty address rejected by check",
			invoke: func(limiter *PerIPRateLimiter, ctx context.Context, address string) error {
				return limiter.CheckLimit(ctx, address)
			},
			wantError: rateLimitTestIPAddressCannotBeEmpty,
		},
		{
			name: "empty address rejected by record",
			invoke: func(limiter *PerIPRateLimiter, ctx context.Context, address string) error {
				return limiter.RecordAttempt(ctx, address)
			},
			wantError: rateLimitTestIPAddressCannotBeEmpty,
		},
		{
			name: rateLimitTestNameCountAttempts,
			configure: func(store *rateLimitErrorStore) {
				store.countErr = errors.New(rateLimitTestInjectedCountFailure)
			},
			invoke: func(limiter *PerIPRateLimiter, ctx context.Context, _ string) error {
				return limiter.CheckLimit(ctx, rateLimitTestIPAddress)
			},
			wantError: rateLimitTestInjectedCountFailure,
		},
		{
			name: rateLimitTestNameRecordAttempt,
			configure: func(store *rateLimitErrorStore) {
				store.recordErr = errors.New(rateLimitTestInjectedRecordFailure)
			},
			invoke: func(limiter *PerIPRateLimiter, ctx context.Context, _ string) error {
				return limiter.RecordAttempt(ctx, rateLimitTestIPAddress)
			},
			wantError: rateLimitTestInjectedRecordFailure,
		},
		{
			name: rateLimitTestNameCleanupExpiredRecords,
			configure: func(store *rateLimitErrorStore) {
				store.cleanupErr = errors.New(rateLimitTestInjectedCleanupFailure)
			},
			invoke: func(limiter *PerIPRateLimiter, ctx context.Context, _ string) error {
				return limiter.Cleanup(ctx)
			},
			wantError: rateLimitTestInjectedCleanupFailure,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := &rateLimitErrorStore{}
			if tc.configure != nil {
				tc.configure(store)
			}

			limiter, err := NewPerIPRateLimiter(store, time.Minute, 1, noop.NewMeterProvider())
			require.NoError(t, err)

			err = tc.invoke(limiter, context.Background(), "")
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestDatabaseRateLimitStore_CleanupDeletesExpiredKey(t *testing.T) {
	t.Parallel()

	store, err := NewDatabaseRateLimitStore(noop.NewMeterProvider())
	require.NoError(t, err)

	now := time.Now().UTC()
	store.nowFn = func() time.Time { return now }
	store.attempts["expired"] = []time.Time{now.Add(-time.Minute)}

	require.NoError(t, store.CleanupExpired(context.Background(), time.Minute))
	require.NotContains(t, store.attempts, "expired")
}

func TestInMemoryRateLimiter_CheckLimitStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*InMemoryRateLimiter, time.Time)
		wantError bool
	}{
		{name: "no attempts"},
		{
			name: "active window below limit",
			configure: func(limiter *InMemoryRateLimiter, now time.Time) {
				limiter.attempts[testUser123] = &attemptRecord{
					count:       limiter.maxAttempts - 1,
					windowStart: now,
				}
			},
		},
		{
			name: "expired window",
			configure: func(limiter *InMemoryRateLimiter, now time.Time) {
				limiter.attempts[testUser123] = &attemptRecord{
					count:       limiter.maxAttempts,
					windowStart: now.Add(-limiter.window - time.Second),
				}
			},
		},
		{
			name: "active window at limit",
			configure: func(limiter *InMemoryRateLimiter, now time.Time) {
				limiter.attempts[testUser123] = &attemptRecord{
					count:       limiter.maxAttempts,
					windowStart: now,
				}
			},
			wantError: true,
		},
		{
			name: "locked out",
			configure: func(limiter *InMemoryRateLimiter, now time.Time) {
				limiter.attempts[testUser123] = &attemptRecord{
					count:       limiter.maxAttempts,
					windowStart: now,
					lockedUntil: now.Add(time.Minute),
				}
			},
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			limiter := NewInMemoryRateLimiter()

			now := time.Now().UTC()
			if tc.configure != nil {
				tc.configure(limiter, now)
			}

			err := limiter.CheckLimit(context.Background(), testUser123)
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInMemoryRateLimiter_RecordAttemptStateTransitions(t *testing.T) {
	t.Parallel()

	limiter := NewInMemoryRateLimiter()
	limiter.maxAttempts = 2
	limiter.window = time.Minute
	now := time.Now().UTC()
	limiter.attempts[testUser123] = &attemptRecord{
		count:       limiter.maxAttempts,
		windowStart: now.Add(-limiter.window - time.Second),
		lockedUntil: now.Add(time.Minute),
	}

	require.NoError(t, limiter.RecordAttempt(context.Background(), testUser123, false))
	record := limiter.attempts[testUser123]
	require.Equal(t, 1, record.count)
	require.WithinDuration(t, now, record.windowStart, time.Second)
	require.True(t, record.lockedUntil.IsZero())

	require.NoError(t, limiter.RecordAttempt(context.Background(), testUser123, false))
	require.False(t, record.lockedUntil.IsZero())
	require.Error(t, limiter.CheckLimit(context.Background(), testUser123))

	require.NoError(t, limiter.RecordAttempt(context.Background(), testUser123, true))
	require.Zero(t, record.count)
	require.True(t, record.lockedUntil.IsZero())
}
