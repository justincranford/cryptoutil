// Copyright (c) 2025-2026 Justin Cranford.
package readiness

import (
	"context"
	crand "crypto/rand"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMonitorLifecycleValidation(t *testing.T) {
	t.Parallel()

	monitor, err := New(Config{
		Checks: []Check{{
			Name:       testDatabaseCheckName,
			ReasonCode: testDatabaseUnavailableReason,
			Run:        func(context.Context) error { return nil },
		}},
	})
	require.NoError(t, err)

	contextForNilArgument := func() context.Context { return nil }
	require.ErrorContains(t, monitor.Start(contextForNilArgument()), "context cannot be nil")
	require.ErrorContains(t, monitor.WaitInitial(contextForNilArgument()), "context cannot be nil")

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	require.NoError(t, monitor.Start(ctx))
	require.ErrorContains(t, monitor.Start(ctx), "already started")
	require.NoError(t, monitor.WaitInitial(ctx))
	require.NoError(t, monitor.Stop(context.Background()))
}

func TestMonitorWaitInitialHonorsCallerDeadline(t *testing.T) {
	t.Parallel()

	checkEntered := make(chan struct{})
	releaseCheck := make(chan struct{})
	monitor, err := New(Config{
		Checks: []Check{{
			Name:       testDatabaseCheckName,
			ReasonCode: testDatabaseUnavailableReason,
			Run: func(context.Context) error {
				close(checkEntered)
				<-releaseCheck

				return nil
			},
		}},
		Period: 10 * time.Second,
		Jitter: func(period time.Duration) time.Duration { return period },
	})
	require.NoError(t, err)

	require.NoError(t, monitor.Start(context.Background()))

	select {
	case <-checkEntered:
	case <-time.After(time.Second):
		close(releaseCheck)
		t.Fatal("readiness check did not start")
	}

	waitCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err = monitor.WaitInitial(waitCtx)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	close(releaseCheck)
	require.NoError(t, monitor.WaitInitial(context.Background()))
	require.NoError(t, monitor.Stop(context.Background()))
}

func TestMonitorStopBeforeStart(t *testing.T) {
	t.Parallel()

	monitor, err := New(Config{
		Checks: []Check{{
			Name:       testDatabaseCheckName,
			ReasonCode: testDatabaseUnavailableReason,
			Run:        func(context.Context) error { return nil },
		}},
	})
	require.NoError(t, err)
	require.NoError(t, monitor.Stop(context.Background()))
}

func TestMonitorStopHonorsCallerDeadline(t *testing.T) {
	t.Parallel()

	checkEntered := make(chan struct{})
	releaseCheck := make(chan struct{})
	monitor, err := New(Config{
		Checks: []Check{{
			Name:       testDatabaseCheckName,
			ReasonCode: testDatabaseUnavailableReason,
			Run: func(context.Context) error {
				close(checkEntered)
				<-releaseCheck

				return nil
			},
		}},
		Period: 10 * time.Second,
		Jitter: func(period time.Duration) time.Duration { return period },
	})
	require.NoError(t, err)
	require.NoError(t, monitor.Start(context.Background()))

	select {
	case <-checkEntered:
	case <-time.After(time.Second):
		close(releaseCheck)
		t.Fatal("readiness check did not start")
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	require.ErrorIs(t, monitor.Stop(stopCtx), context.DeadlineExceeded)

	close(releaseCheck)
	require.NoError(t, monitor.Stop(context.Background()))
}

func TestMonitorUsesDefaultConfigurationAndJitter(t *testing.T) {
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
	})
	require.NoError(t, err)
	require.Equal(t, defaultPeriod, monitor.config.Period)
	require.Equal(t, defaultCheckTimeout, monitor.config.CheckTimeout)
	require.Equal(t, defaultMaxStaleness, monitor.config.MaxStaleness)

	require.NoError(t, monitor.Start(context.Background()))
	require.NoError(t, monitor.WaitInitial(context.Background()))
	require.GreaterOrEqual(t, callCount.Load(), int64(1))
	require.NoError(t, monitor.Stop(context.Background()))
}

func TestMonitorNormalizesNonPositiveJitterAndRepeatsChecks(t *testing.T) {
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
		Period: time.Second,
		Jitter: func(time.Duration) time.Duration { return 0 },
	})
	require.NoError(t, err)
	require.NoError(t, monitor.Start(context.Background()))
	require.NoError(t, monitor.WaitInitial(context.Background()))
	require.Eventually(t, func() bool {
		return callCount.Load() >= 2
	}, time.Second, time.Millisecond)
	require.NoError(t, monitor.Stop(context.Background()))
}

func TestMonitorSortsChecksDeterministically(t *testing.T) {
	t.Parallel()

	monitor, err := New(Config{
		Checks: []Check{
			{
				Name:       "zeta",
				ReasonCode: "zeta_unavailable",
				Run:        func(context.Context) error { return nil },
			},
			{
				Name:       "alpha",
				ReasonCode: "alpha_unavailable",
				Run:        func(context.Context) error { return nil },
			},
		},
		Period: 10 * time.Second,
		Jitter: func(period time.Duration) time.Duration { return period },
	})
	require.NoError(t, err)
	require.Equal(t, "alpha", monitor.config.Checks[0].Name)
	require.Equal(t, "zeta", monitor.config.Checks[1].Name)

	require.NoError(t, monitor.Start(context.Background()))
	require.NoError(t, monitor.WaitInitial(context.Background()))
	snapshot := monitor.Snapshot()
	require.Len(t, snapshot.Checks, 2)
	require.Equal(t, "alpha", snapshot.Checks[0].Name)
	require.Equal(t, "zeta", snapshot.Checks[1].Name)
	require.NoError(t, monitor.Stop(context.Background()))
}

func TestJitteredInterval(t *testing.T) {
	t.Parallel()

	const (
		period      = 10 * time.Second
		smallPeriod = 5 * time.Nanosecond
	)

	tests := []struct {
		name         string
		period       time.Duration
		randomReader *strings.Reader
		wantRandom   bool
	}{
		{
			name:         "randomized interval stays within bounds",
			period:       period,
			randomReader: nil,
			wantRandom:   true,
		},
		{
			name:         "small period is unchanged",
			period:       smallPeriod,
			randomReader: strings.NewReader(""),
		},
		{
			name:         "random source error uses the base period",
			period:       period,
			randomReader: strings.NewReader(""),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			reader := io.Reader(crand.Reader)
			if testCase.randomReader != nil {
				reader = testCase.randomReader
			}

			got := jitteredIntervalWithReader(testCase.period, reader)
			if testCase.wantRandom {
				require.GreaterOrEqual(t, got, testCase.period-testCase.period/jitterWindowDivisor)
				require.LessOrEqual(t, got, testCase.period+testCase.period/jitterWindowDivisor)

				return
			}

			require.Equal(t, testCase.period, got)
		})
	}
}

func TestStopAndDrainTimerAfterTimerWasRead(t *testing.T) {
	t.Parallel()

	timer := time.NewTimer(10 * time.Millisecond)
	select {
	case <-timer.C:
	case <-time.After(time.Second):
		t.Fatal("timer did not expire")
	}

	require.False(t, timer.Stop())

	stopped := make(chan struct{})

	go func() {
		stopAndDrainTimer(timer)
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("stopping an expired timer blocked while draining")
	}
}
