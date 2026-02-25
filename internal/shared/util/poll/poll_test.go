// Copyright (c) 2025 Justin Cranford
//
//

package poll_test

import (
	"context"
	"errors"
	"testing"
	"time"

	cryptoutilSharedUtilPoll "cryptoutil/internal/shared/util/poll"

	"github.com/stretchr/testify/require"
)

func TestUntil(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		timeout      time.Duration
		interval     time.Duration
		condFn       func() cryptoutilSharedUtilPoll.ConditionFunc
		ctxFn        func() context.Context
		wantErr      bool
		wantSentinel error
		errContains  string
	}{
		{
			name:     "immediate success",
			timeout:  1 * time.Second,
			interval: 10 * time.Millisecond,
			condFn: func() cryptoutilSharedUtilPoll.ConditionFunc {
				return func(_ context.Context) (bool, error) {
					return true, nil
				}
			},
			ctxFn:   func() context.Context { return context.Background() },
			wantErr: false,
		},
		{
			name:     "success after retries",
			timeout:  1 * time.Second,
			interval: 10 * time.Millisecond,
			condFn: func() cryptoutilSharedUtilPoll.ConditionFunc {
				callCount := 0

				return func(_ context.Context) (bool, error) {
					callCount++

					return callCount >= 3, nil
				}
			},
			ctxFn:   func() context.Context { return context.Background() },
			wantErr: false,
		},
		{
			name:     "timeout exceeded",
			timeout:  50 * time.Millisecond,
			interval: 20 * time.Millisecond,
			condFn: func() cryptoutilSharedUtilPoll.ConditionFunc {
				return func(_ context.Context) (bool, error) {
					return false, nil
				}
			},
			ctxFn:        func() context.Context { return context.Background() },
			wantErr:      true,
			wantSentinel: cryptoutilSharedUtilPoll.ErrTimeout,
			errContains:  "poll timed out",
		},
		{
			name:     "fatal error stops polling",
			timeout:  1 * time.Second,
			interval: 10 * time.Millisecond,
			condFn: func() cryptoutilSharedUtilPoll.ConditionFunc {
				return func(_ context.Context) (bool, error) {
					return false, errors.New("fatal check error")
				}
			},
			ctxFn:       func() context.Context { return context.Background() },
			wantErr:     true,
			errContains: "poll condition failed",
		},
		{
			name:     "context canceled before first call",
			timeout:  5 * time.Second,
			interval: 10 * time.Millisecond,
			condFn: func() cryptoutilSharedUtilPoll.ConditionFunc {
				return func(_ context.Context) (bool, error) {
					return false, nil
				}
			},
			ctxFn: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			},
			wantErr:     true,
			errContains: "poll canceled",
		},
		{
			name:        "nil conditionFn",
			timeout:     1 * time.Second,
			interval:    10 * time.Millisecond,
			condFn:      func() cryptoutilSharedUtilPoll.ConditionFunc { return nil },
			ctxFn:       func() context.Context { return context.Background() },
			wantErr:     true,
			errContains: "conditionFn must not be nil",
		},
		{
			name:     "zero timeout",
			timeout:  0,
			interval: 10 * time.Millisecond,
			condFn: func() cryptoutilSharedUtilPoll.ConditionFunc {
				return func(_ context.Context) (bool, error) {
					return true, nil
				}
			},
			ctxFn:       func() context.Context { return context.Background() },
			wantErr:     true,
			errContains: "poll timeout must be positive",
		},
		{
			name:     "negative timeout",
			timeout:  -1 * time.Second,
			interval: 10 * time.Millisecond,
			condFn: func() cryptoutilSharedUtilPoll.ConditionFunc {
				return func(_ context.Context) (bool, error) {
					return true, nil
				}
			},
			ctxFn:       func() context.Context { return context.Background() },
			wantErr:     true,
			errContains: "poll timeout must be positive",
		},
		{
			name:     "zero interval",
			timeout:  1 * time.Second,
			interval: 0,
			condFn: func() cryptoutilSharedUtilPoll.ConditionFunc {
				return func(_ context.Context) (bool, error) {
					return true, nil
				}
			},
			ctxFn:       func() context.Context { return context.Background() },
			wantErr:     true,
			errContains: "poll interval must be positive",
		},
		{
			name:     "negative interval",
			timeout:  1 * time.Second,
			interval: -1 * time.Millisecond,
			condFn: func() cryptoutilSharedUtilPoll.ConditionFunc {
				return func(_ context.Context) (bool, error) {
					return true, nil
				}
			},
			ctxFn:       func() context.Context { return context.Background() },
			wantErr:     true,
			errContains: "poll interval must be positive",
		},
		{
			name:     "context canceled during polling",
			timeout:  5 * time.Second,
			interval: 10 * time.Millisecond,
			condFn: func() cryptoutilSharedUtilPoll.ConditionFunc {
				return func(_ context.Context) (bool, error) {
					return false, nil
				}
			},
			ctxFn: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
				_ = cancel // let timeout handle cancellation.

				return ctx
			},
			wantErr:     true,
			errContains: "poll canceled",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := cryptoutilSharedUtilPoll.Until(tc.ctxFn(), tc.timeout, tc.interval, tc.condFn())

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)

				if tc.wantSentinel != nil {
					require.ErrorIs(t, err, tc.wantSentinel)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
