// Copyright (c) 2025-2026 Justin Cranford.
package readiness

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMonitorWaitReadyRetainsFailClosedStartup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		recover bool
	}{
		{name: "transient initial failure recovers", recover: true},
		{name: "persistent failure honors caller deadline"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var healthy atomic.Bool
			monitor, err := New(Config{
				Period: time.Millisecond,
				Jitter: func(period time.Duration) time.Duration { return period },
				Checks: []Check{{
					Name: testDatabaseCheckName, ReasonCode: testDatabaseUnavailableReason,
					Run: func(context.Context) error {
						if healthy.Load() {
							return nil
						}

						return ErrInitialChecksFailed
					},
				}},
			})
			require.NoError(t, err)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			require.NoError(t, monitor.Start(ctx))
			defer func() { require.NoError(t, monitor.Stop(context.Background())) }()
			require.ErrorIs(t, monitor.WaitInitial(ctx), ErrInitialChecksFailed)
			require.False(t, monitor.Snapshot().Ready)

			healthy.Store(tc.recover)
			waitCtx, waitCancel := context.WithTimeout(ctx, time.Second)
			if !tc.recover {
				waitCancel()
			}
			defer waitCancel()

			err = monitor.WaitReady(waitCtx)
			if tc.recover {
				require.NoError(t, err)
				require.True(t, monitor.Snapshot().Ready)
			} else {
				require.ErrorIs(t, err, context.Canceled)
				require.False(t, monitor.Snapshot().Ready)
			}
		})
	}
}

func TestMonitorWaitReadyValidatesContextAndDeadline(t *testing.T) {
	t.Parallel()

	monitor, err := New(Config{Checks: []Check{{
		Name: testDatabaseCheckName, ReasonCode: testDatabaseUnavailableReason,
		Run: func(context.Context) error { return nil },
	}}})
	require.NoError(t, err)
	nilContext := func() context.Context { return nil }
	require.ErrorContains(t, monitor.WaitReady(nilContext()), "context cannot be nil")
	ctx, cancel := context.WithTimeout(t.Context(), time.Millisecond)
	defer cancel()
	require.ErrorIs(t, monitor.WaitReady(ctx), context.DeadlineExceeded)
}
