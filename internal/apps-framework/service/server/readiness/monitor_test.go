// Copyright (c) 2025-2026 Justin Cranford.
package readiness

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testDatabaseCheckName         = cryptoutilSharedMagic.RealmStorageTypeDatabase
	testDatabaseUnavailableReason = "database_unavailable"
)

func TestMonitorInitialObservationAndStop(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int64

	monitor, err := New(Config{
		Checks: []Check{{
			Name:       testDatabaseCheckName,
			ReasonCode: testDatabaseUnavailableReason,
			Run: func(context.Context) error {
				callCount.Add(1)

				return nil
			},
		}},
		Period: 10 * time.Second,
		Jitter: func(period time.Duration) time.Duration { return period },
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	require.NoError(t, monitor.Start(ctx))
	require.NoError(t, monitor.WaitInitial(ctx))

	snapshot := monitor.Snapshot()
	require.True(t, snapshot.Ready)
	require.Len(t, snapshot.Checks, 1)
	require.Equal(t, testDatabaseCheckName, snapshot.Checks[0].Name)
	require.GreaterOrEqual(t, callCount.Load(), int64(1))

	require.NoError(t, monitor.Stop(context.Background()))
}

func TestMonitorInitialFailureSanitizesCheckError(t *testing.T) {
	t.Parallel()

	monitor, err := New(Config{
		Checks: []Check{{
			Name:       testDatabaseCheckName,
			ReasonCode: testDatabaseUnavailableReason,
			Run: func(context.Context) error {
				return errors.New("postgres://user:secret@db/password")
			},
		}},
		Period: 10 * time.Second,
		Jitter: func(period time.Duration) time.Duration { return period },
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	require.NoError(t, monitor.Start(ctx))
	err = monitor.WaitInitial(ctx)

	require.ErrorIs(t, err, ErrInitialChecksFailed)
	require.Contains(t, err.Error(), testDatabaseUnavailableReason)
	require.NotContains(t, err.Error(), "secret")
	require.False(t, monitor.Snapshot().Ready)
	require.NoError(t, monitor.Stop(context.Background()))
}

func TestMonitorSnapshotFailsClosedWhenStale(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	monitor, err := New(Config{
		Checks: []Check{{
			Name:       "collector",
			ReasonCode: "collector_unavailable",
			Run:        func(context.Context) error { return nil },
		}},
		Period:       10 * time.Second,
		MaxStaleness: 5 * time.Second,
		Now:          func() time.Time { return now },
		Jitter:       func(period time.Duration) time.Duration { return period },
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	require.NoError(t, monitor.Start(ctx))
	require.NoError(t, monitor.WaitInitial(ctx))

	now = now.Add(6 * time.Second)
	snapshot := monitor.Snapshot()
	require.False(t, snapshot.Ready)
	require.True(t, snapshot.Checks[0].Stale)
	require.Equal(t, 6*time.Second, snapshot.Checks[0].Age)
	require.NoError(t, monitor.Stop(context.Background()))
}

func TestMonitorCheckTimeout(t *testing.T) {
	t.Parallel()

	monitor, err := New(Config{
		Checks: []Check{{
			Name:       testDatabaseCheckName,
			ReasonCode: testDatabaseUnavailableReason,
			Run: func(ctx context.Context) error {
				<-ctx.Done()

				return ctx.Err()
			},
		}},
		CheckTimeout: 10 * time.Millisecond,
		Period:       10 * time.Second,
		Jitter:       func(period time.Duration) time.Duration { return period },
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	require.NoError(t, monitor.Start(ctx))
	err = monitor.WaitInitial(ctx)
	require.ErrorIs(t, err, ErrInitialChecksFailed)
	require.Contains(t, err.Error(), testDatabaseUnavailableReason)
	require.NoError(t, monitor.Stop(context.Background()))
}

func TestNewRejectsInvalidChecks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config Config
		want   string
	}{
		{
			name: "no checks",
			want: "at least one readiness check",
		},
		{
			name: "unsanitized check name",
			config: Config{Checks: []Check{{
				Name:       "database/password",
				ReasonCode: testDatabaseUnavailableReason,
				Run:        func(context.Context) error { return nil },
			}}},
			want: "sanitized identifiers",
		},
		{
			name: "missing check function",
			config: Config{Checks: []Check{{
				Name:       testDatabaseCheckName,
				ReasonCode: testDatabaseUnavailableReason,
			}}},
			want: "has no function",
		},
		{
			name: "duplicate check name",
			config: Config{Checks: []Check{
				{Name: testDatabaseCheckName, ReasonCode: testDatabaseUnavailableReason, Run: func(context.Context) error { return nil }},
				{Name: testDatabaseCheckName, ReasonCode: testDatabaseUnavailableReason, Run: func(context.Context) error { return nil }},
			}},
			want: "duplicate readiness check",
		},
		{
			name: "non-positive interval",
			config: Config{
				Checks:       []Check{{Name: testDatabaseCheckName, ReasonCode: testDatabaseUnavailableReason, Run: func(context.Context) error { return nil }}},
				Period:       -time.Second,
				CheckTimeout: time.Second,
				MaxStaleness: time.Second,
			},
			want: "durations must be positive",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(testCase.config)
			require.ErrorContains(t, err, testCase.want)
		})
	}
}
